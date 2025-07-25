package v2

//go:generate mockgen -destination=../../../../mocks/services/oauthV2/mock_roundtripper.go -package=mock_oauthV2 net/http RoundTripper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats"
	kitsync "github.com/rudderlabs/rudder-go-kit/sync"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
	backendconfig "github.com/rudderlabs/rudder-server/backend-config"
	oauth "github.com/rudderlabs/rudder-server/services/oauth/v2"
	"github.com/rudderlabs/rudder-server/services/oauth/v2/common"
	cntx "github.com/rudderlabs/rudder-server/services/oauth/v2/context"
	oauthexts "github.com/rudderlabs/rudder-server/services/oauth/v2/extensions"
)

// TransportArgs is a struct that contains the required parameters to create a new Oauth2Transport.
type TransportArgs struct {
	BackendConfig backendconfig.BackendConfig
	FlowType      common.RudderFlow
	// TokenCache is a cache for storing OAuth tokens.
	TokenCache *oauth.Cache
	// Locker provides synchronization mechanisms.
	Locker *kitsync.PartitionRWLocker
	// GetAuthErrorCategory is a function to get the auth error category from the response body. It can be REFRESH_TOKEN or AUTH_STATUS_INACTIVE.
	GetAuthErrorCategory func([]byte) (string, error)
	// Augmenter is an interface for augmenting requests with OAuth tokens.
	oauthexts.Augmenter
	// OAuthHandler handles refreshToken and fetchToken requests.
	OAuthHandler *oauth.OAuthHandler
	// OriginalTransport is the underlying HTTP transport.
	OriginalTransport http.RoundTripper
	logger            logger.Logger
	stats             stats.Stats
}

// OAuthTransport is a http.RoundTripper that adds the appropriate authorization information to oauth requests.
// Also, it makes the calls to the actual endpoint and handles the response by refreshing the token if required or by updating the authStatus to "inactive".
type OAuthTransport struct {
	oauthHandler oauth.OAuthHandler
	oauthexts.Augmenter
	Transport            http.RoundTripper
	log                  logger.Logger
	flow                 common.RudderFlow
	getAuthErrorCategory func([]byte) (string, error)
	stats                stats.Stats
}

// This struct is used to transport common information across the pre and post round trip methods.
type roundTripState struct {
	destination        *oauth.DestinationInfo
	accountID          string
	refreshTokenParams *oauth.RefreshTokenParams
	res                *http.Response
	req                *http.Request
}

func httpResponseCreator(statusCode int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{"apiVersion": []string{"2"}},
	}
}

func NewOAuthTransport(args *TransportArgs) *OAuthTransport {
	t := &OAuthTransport{
		oauthHandler:         *args.OAuthHandler,
		Augmenter:            args.Augmenter,
		Transport:            args.OriginalTransport,
		log:                  args.logger,
		flow:                 args.FlowType,
		getAuthErrorCategory: args.GetAuthErrorCategory,
		stats:                args.stats,
	}
	if t.log == nil {
		t.log = logger.NewLogger()
	}
	if t.stats == nil {
		t.stats = stats.Default
	}
	t.log = t.log.Child("OAuthHttpTransport")
	return t
}

func (t *OAuthTransport) preRoundTrip(rts *roundTripState) *http.Response {
	if t.Augmenter == nil {
		return nil
	}
	body, err := io.ReadAll(rts.req.Body)
	if err != nil {
		t.log.Errorn("[OAuthPlatformError] reading request body",
			obskit.DestinationID(rts.destination.ID),
			obskit.WorkspaceID(rts.destination.WorkspaceID),
			obskit.DestinationType(rts.destination.DefinitionName),
			logger.NewStringField("flow", string(t.flow)))
		return httpResponseCreator(http.StatusInternalServerError, []byte(fmt.Errorf("reading request body pre roundTrip: %w", err).Error()))
	}
	statusCode, authResponse, err := t.oauthHandler.FetchToken(rts.refreshTokenParams)
	if statusCode == http.StatusOK {
		rts.req = rts.req.WithContext(cntx.CtxWithSecret(rts.req.Context(), authResponse.Account.Secret))
		err = t.Augment(rts.req, body, authResponse.Account.Secret)
		if err != nil {
			t.log.Errorn("[OAuthPlatformError] secret augmentation",
				obskit.DestinationID(rts.destination.ID),
				obskit.WorkspaceID(rts.destination.WorkspaceID),
				obskit.DestinationType(rts.destination.DefinitionName),
				logger.NewStringField("flow", string(t.flow)),
				logger.NewErrorField(err))
			return httpResponseCreator(http.StatusInternalServerError, []byte(fmt.Errorf("augmenting the secret pre roundTrip: %w", err).Error()))
		}
		return nil
	} else if authResponse != nil && authResponse.Err == common.RefTokenInvalidGrant {
		t.oauthHandler.AuthStatusToggle(&oauth.AuthStatusToggleParams{
			Destination:     rts.destination,
			WorkspaceID:     rts.destination.WorkspaceID,
			RudderAccountID: rts.accountID,
			StatPrefix:      common.AuthStatusInActive,
			AuthStatus:      common.CategoryAuthStatusInactive,
		})
		return httpResponseCreator(http.StatusBadRequest, []byte(err.Error()))
	}
	return httpResponseCreator(statusCode, []byte(err.Error()))
}

func (t *OAuthTransport) postRoundTrip(rts *roundTripState) *http.Response {
	respData, err := io.ReadAll(rts.res.Body)
	if err != nil {
		t.log.Errorn("[OAuthPlatformError] reading response body",
			obskit.DestinationID(rts.destination.ID),
			obskit.WorkspaceID(rts.destination.WorkspaceID),
			obskit.DestinationType(rts.destination.DefinitionName),
			logger.NewStringField("flow", string(t.flow)),
			logger.NewErrorField(err),
		)
		// Create a new response with a 500 status code
		return httpResponseCreator(http.StatusInternalServerError, []byte(fmt.Sprintf("[postRoundTrip]Error reading response body: %v", err)))
	}
	interceptorResp := oauth.OAuthInterceptorResponse{}
	// internal function
	applyInterceptorRespToHttpResp := func() {
		var interceptorRespBytes []byte
		transResp := oauth.TransportResponse{
			OriginalResponse:    string(respData),
			InterceptorResponse: interceptorResp,
		}
		interceptorRespBytes, _ = jsonrs.Marshal(transResp)
		rts.res.Body = io.NopCloser(bytes.NewReader(interceptorRespBytes))
	}
	authErrorCategory, err := t.getAuthErrorCategory(respData)
	if err != nil {
		t.log.Errorn("[OAuthPlatformError] get authErrorCategory",
			obskit.DestinationID(rts.destination.ID),
			obskit.WorkspaceID(rts.destination.WorkspaceID),
			obskit.DestinationType(rts.destination.DefinitionName),
			logger.NewStringField("flow", string(t.flow)),
			logger.NewErrorField(errors.New(string(respData))),
		)
		// Instead of returning an error, set a 500 status code in the interceptor response
		// This will make the error retryable instead of causing a panic
		interceptorResp.StatusCode = http.StatusInternalServerError
		interceptorResp.Response = "[OAuthPlatformError]Error processing response: " + string(respData)
		applyInterceptorRespToHttpResp()
		return rts.res
	}

	switch authErrorCategory {
	case common.CategoryRefreshToken:
		// since same token that was used to make the http call needs to be refreshed, we need the current token information
		var oldSecret json.RawMessage
		oldSecret, ok := cntx.SecretFromCtx(rts.req.Context())
		if !ok {
			t.log.Errorn("[OAuthPlatformError] get secret from context",
				obskit.DestinationID(rts.destination.ID),
				obskit.WorkspaceID(rts.destination.WorkspaceID),
				obskit.DestinationType(rts.destination.DefinitionName),
				logger.NewStringField("flow", string(t.flow)))
			// Instead of returning an error, set a 500 status code in the interceptor response
			interceptorResp.StatusCode = http.StatusInternalServerError
			interceptorResp.Response = "[OAuthPlatformError]Error getting secret from context"
			applyInterceptorRespToHttpResp()
			return rts.res
		}
		rts.refreshTokenParams.Secret = oldSecret
		rts.refreshTokenParams.DestinationID = rts.destination.ID
		_, authResponse, refErr := t.oauthHandler.RefreshToken(rts.refreshTokenParams)
		if refErr != nil {
			interceptorResp.Response = refErr.Error()
		}
		// refresh token success(retry)
		// refresh token failed
		// It can be failed due to the following reasons
		// 1. invalid grant(abort)
		// 2. control plan api call failed(retry)
		// 3. some error happened while returning from RefreshToken function(retry)
		if authResponse != nil && authResponse.Err == common.RefTokenInvalidGrant {
			// Setting the response we obtained from trying to Refresh the token
			interceptorResp.StatusCode = http.StatusBadRequest
			t.oauthHandler.AuthStatusToggle(&oauth.AuthStatusToggleParams{
				Destination:     rts.destination,
				WorkspaceID:     rts.destination.WorkspaceID,
				RudderAccountID: rts.accountID,
				StatPrefix:      common.AuthStatusInActive,
				AuthStatus:      common.AuthStatusInActive,
			})
			// rts.res.Body = errorInRefToken
			interceptorResp.Response = authResponse.ErrorMessage
			applyInterceptorRespToHttpResp()
			return rts.res
		}
		interceptorResp.StatusCode = http.StatusInternalServerError
	case common.CategoryAuthStatusInactive:
		t.oauthHandler.AuthStatusToggle(&oauth.AuthStatusToggleParams{
			Destination:     rts.destination,
			WorkspaceID:     rts.destination.WorkspaceID,
			RudderAccountID: rts.accountID,
			StatPrefix:      common.AuthStatusInActive,
			AuthStatus:      common.AuthStatusInActive,
		})
		interceptorResp.StatusCode = http.StatusBadRequest
	}
	applyInterceptorRespToHttpResp()
	// when error is not nil, the response sent will be ignored(downstream)
	return rts.res
}

func (rts *roundTripState) getAccountID(flow common.RudderFlow) (string, error) {
	accountId, err := rts.destination.GetAccountID(flow)
	if err != nil {
		return "", fmt.Errorf("[OAuthPlatformError]accountId not found for destination(%s) in %s flow", rts.destination.ID, flow)
	}
	if accountId == "" {
		return "", fmt.Errorf("[OAuthPlatformError]accountId is empty for destination(%s) in %s flow", rts.destination.ID, flow)
	}
	return accountId, nil
}

func (t *OAuthTransport) fireTimerStats(statName string, tags stats.Tags, startTime time.Time) {
	t.stats.NewTaggedStat(statName, stats.TimerType, tags).SendTiming(time.Since(startTime))
}

func (t *OAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	destination, ok := cntx.DestInfoFromCtx(req.Context())
	if !ok || destination == nil {
		t.log.Errorn("[OAuthPlatformError]destinationInfo is not present in request context", logger.NewStringField("flow", string(t.flow)))
		return httpResponseCreator(http.StatusInternalServerError, []byte("[OAuthPlatformError]request context data is not of destinationInfo type")), nil
	}

	tags := stats.Tags{
		"flow":           string(t.flow),
		"destinationId":  destination.ID,
		"workspaceId":    destination.WorkspaceID,
		"destType":       destination.DefinitionName,
		"origRequestURL": req.URL.Path,
	}
	startTime := time.Now()
	defer t.fireTimerStats("oauth_v2_http_total_roundtrip_latency", tags, startTime)
	isOauthDestination, err := destination.IsOAuthDestination(t.flow)
	if err != nil {
		t.log.Errorn("[OAuthPlatformError]checking if destination is oauth destination",
			obskit.DestinationID(destination.ID),
			obskit.WorkspaceID(destination.WorkspaceID),
			obskit.DestinationType(destination.DefinitionName),
			logger.NewStringField("flow", string(t.flow)),
			logger.NewErrorField(err),
		)
		return httpResponseCreator(http.StatusInternalServerError, []byte(fmt.Sprintf("[OAuthPlatformError]checking if destination is oauth destination: %v", err.Error()))), nil
	}
	if !isOauthDestination {
		return t.Transport.RoundTrip(req)
	}
	rts := &roundTripState{}
	rts.destination = destination
	rts.accountID, err = rts.getAccountID(t.flow)
	if rts.accountID == "" {
		t.log.Errorn("[OAuthPlatformError]accountId not found or empty for destination",
			obskit.DestinationID(rts.destination.ID),
			obskit.WorkspaceID(rts.destination.WorkspaceID),
			obskit.DestinationType(rts.destination.DefinitionName),
			logger.NewStringField("flow", string(t.flow)),
			logger.NewErrorField(err),
		)
		return httpResponseCreator(http.StatusInternalServerError, []byte(err.Error())), nil
	}
	rts.refreshTokenParams = &oauth.RefreshTokenParams{
		AccountID:     rts.accountID,
		WorkspaceID:   rts.destination.WorkspaceID,
		DestDefName:   rts.destination.DefinitionName,
		DestinationID: rts.destination.ID,
	}
	rts.req = req
	preRoundTripStartTime := time.Now()
	errorHttpResponse := t.preRoundTrip(rts)
	t.fireTimerStats("oauth_v2_http_pre_roundtrip_latency", tags, preRoundTripStartTime)
	if errorHttpResponse != nil {
		return errorHttpResponse, nil
	}
	roundTripStartTime := time.Now()
	res, err := t.Transport.RoundTrip(rts.req)
	if err != nil {
		t.log.Errorn("[RoundTrip]transport round trip",
			obskit.DestinationID(rts.destination.ID),
			obskit.WorkspaceID(rts.destination.WorkspaceID),
			obskit.DestinationType(rts.destination.DefinitionName),
			logger.NewStringField("flow", string(t.flow)),
			logger.NewErrorField(err))
		// Return a 500 error response instead of propagating the error
		return httpResponseCreator(http.StatusInternalServerError, []byte(fmt.Sprintf("Transport round trip error: %v", err))), nil
	}
	t.fireTimerStats("oauth_v2_http_roundtrip_latency", tags, roundTripStartTime)
	rts.res = res
	postRoundTripStartTime := time.Now()
	defer t.fireTimerStats("oauth_v2_http_post_roundtrip_latency", tags, postRoundTripStartTime)
	return t.postRoundTrip(rts), nil
}
