package authhandler

import (
	"authorization_flow_oauth/internal/config"
	"authorization_flow_oauth/internal/store"
	"authorization_flow_oauth/internal/utils"
	"authorization_flow_oauth/pkg/auth"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg          *config.Config
	serverAddr   string
	authClient   *auth.Client
	authStore    store.AuthStore
	sessionStore store.SessionStore
}

func New(
	cfg *config.Config,
	serverAddr string,
	authClient *auth.Client,
	authStore store.AuthStore,
	sessionStore store.SessionStore,
) *AuthHandler {
	return &AuthHandler{
		cfg:          cfg,
		serverAddr:   serverAddr,
		authClient:   authClient,
		authStore:    authStore,
		sessionStore: sessionStore,
	}
}

func (a *AuthHandler) RedirectToKeycloak(c *gin.Context) {
	stateID, err := utils.GenerateRandomBase64Str()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	if err = a.authStore.SetState(c, stateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.Redirect(http.StatusFound, a.authClient.Oauth.AuthCodeURL(stateID))
}

func (a *AuthHandler) RenderLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login/login.tmpl", gin.H{
		"Title":       "Welcome",
		"Message":     "Please log in to continue",
		"KeycloakURL": fmt.Sprintf("http://%s/login-keycloak", a.serverAddr),
	})
}

func (a *AuthHandler) LogoutHandler(c *gin.Context) {

	idToken, err := c.Cookie("idToken")
	if err != nil || idToken == "" {
		c.Redirect(http.StatusFound, "/")
		c.Abort()
		return
	}

    logoutURL := fmt.Sprintf("%s/protocol/openid-connect/logout",
        a.authClient.Provider.Endpoint().AuthURL[:strings.Index(a.authClient.Provider.Endpoint().AuthURL, "/protocol")],
    )

    params := url.Values{}
	params.Set("client_id", a.authClient.Oauth.ClientID)
    params.Set("id_token_hint", idToken)
    params.Set("post_logout_redirect_uri", "http://localhost:8081")
    finalURL := fmt.Sprintf("%s?%s", logoutURL, params.Encode())

    c.Redirect(http.StatusFound, finalURL)
}

