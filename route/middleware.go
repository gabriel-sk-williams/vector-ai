package route

import (
	"errors"
	"net/http"
	"vector-ai/drive"
	"vector-ai/model"
	"vector-ai/util"

	"time"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"goyave.dev/goyave/v4"
)

func (h Handler) Authorization(next goyave.Handler) goyave.Handler {

	return func(res *goyave.Response, req *goyave.Request) {

		claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
		userId := claims.Subject
		orgId := req.Params["orgId"]
		method := req.Method()
		if orgId == "SYSTEM" && method == "GET" {
			next(res, req)
			return
		}

		if orgId != "" { // there is an orgId
			_, err := h.PG.GetUserOrgAssociation(orgId, userId)

			if err != nil {
				res.Status(http.StatusNotFound)
				res.Error(err)
				return
			}
		}

		next(res, req) // Pass to the next handler
	}
}

func (h Handler) Admin(next goyave.Handler) goyave.Handler {
	return func(res *goyave.Response, req *goyave.Request) {

		claims := req.Extra["jwt_claims"].(*model.ClerkClaims)
		userId := claims.Subject

		orgs, err := h.PG.ListOrgsByUserId(userId)

		isAdmin := util.IsAdmin(orgs)

		if isAdmin && err == nil {
			next(res, req) // Pass to the next handler
		} else {
			res.Status(http.StatusNotFound)
			res.Error(err)
			return
		}
	}
}

func (h Handler) Invite(next goyave.Handler) goyave.Handler {
	return func(res *goyave.Response, req *goyave.Request) {

		userId := req.Params["userId"]
		inviteToken := req.String("token")

		// check for active, non-expired invite in the database
		invite, err := h.PG.GetInvite(inviteToken)
		if err != nil {
			res.Status(http.StatusBadRequest)
			res.Error(err)
			return
		}
		now := time.Now()
		expiry := invite.Expiration
		notExpired := now.Before(expiry)

		// check that userId has same email as on invite
		user, err := h.PG.GetUser(userId)
		correctEmail := invite.Email == user.Email

		if correctEmail && notExpired && err == nil {
			next(res, req) // Pass to the next handler
		} else {
			res.Status(http.StatusBadRequest)
			res.Error(err)
			return
		}
	}
}

// oAuth
func (h Handler) checkToken(userId string, provider string) (*clerk.UserOAuthAccessToken, int, error) {

	var token *clerk.UserOAuthAccessToken

	tokens, err := h.CL.Users().ListOAuthAccessTokens(userId, provider)
	if err != nil {
		return token, http.StatusForbidden, err
	}

	if len(tokens) == 0 { // no Google Account connected
		return token, http.StatusUnauthorized, errors.New("no Google Account connected")
	}

	token = tokens[0]

	scopes := token.Scopes
	if !util.Contains(scopes, drive.DriveReadonlyScope) {
		return token, http.StatusForbidden, errors.New("readonly scope is not permitted")
	}

	return token, http.StatusAccepted, nil
}
