package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/affanahmad024/quiz-craft/backend/auth/core"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/user/model"
)

// Signup handles user registration
func Signup(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	var req model.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.Signup::BindErr [ip=%s [err=%v", ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.Signup [email=%s [name=%s", req.Email, req.Name)

	resp, err := core.Signup(ctx, req)
	if err != nil {
		commons.LogC(ctx, "[api.Signup::KO [time_taken=%v [ip=%s [email=%s [err=%v",
			time.Since(st), ip, req.Email, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsConflict(err) {
			commons.Resp409(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.Signup::OK [time_taken=%v [ip=%s [user_id=%s [email=%s",
		time.Since(st), ip, resp.UserID, req.Email)
	commons.RespOKWithMsg(c, "Signup successful. Please check your email to verify your account.", resp)
}

// VerifyEmail verifies user email with token
func VerifyEmail(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	var req model.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.VerifyEmail::BindErr [ip=%s [err=%v", ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.VerifyEmail [user_id=%s", req.UserID)

	err := core.VerifyEmail(ctx, req)
	if err != nil {
		commons.LogC(ctx, "[api.VerifyEmail::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, req.UserID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.VerifyEmail::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, req.UserID)
	commons.RespOKWithMsg(c, "Email verified successfully", nil)
}

// ResendVerification resends verification email
func ResendVerification(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	var req model.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.ResendVerification::BindErr [ip=%s [err=%v", ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.ResendVerification [email=%s", req.Email)

	err := core.ResendVerification(ctx, req)
	if err != nil {
		commons.LogC(ctx, "[api.ResendVerification::KO [time_taken=%v [ip=%s [email=%s [err=%v",
			time.Since(st), ip, req.Email, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ResendVerification::OK [time_taken=%v [ip=%s [email=%s",
		time.Since(st), ip, req.Email)
	commons.RespOKWithMsg(c, "Verification email sent successfully", nil)
}

// Login handles user authentication
func Login(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.Login::BindErr [ip=%s [err=%v", ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.Login [email=%s [ip=%s", req.Email, ip)

	resp, err := core.Login(ctx, req)
	if err != nil {
		commons.LogC(ctx, "[api.Login::KO [time_taken=%v [ip=%s [email=%s [err=%v",
			time.Since(st), ip, req.Email, err)

		if commons.IsUnauthorized(err) {
			commons.Resp401(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	// Set secure cookies
	c.SetCookie(
		"access_token",
		resp.AccessToken,
		int(resp.ExpiresIn),
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // HttpOnly
	)

	c.SetCookie(
		"refresh_token",
		resp.RefreshToken,
		int(7*24*3600), // 7 days
		"/",
		"",
		false,
		true,
	)

	commons.Log(ctx, "[api.Login::OK [time_taken=%v [ip=%s [user_id=%s [email=%s",
		time.Since(st), ip, resp.User.ID, req.Email)
	commons.RespOKWithMsg(c, "Login successful", resp)
}

// RefreshToken refreshes access token
func RefreshToken(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.RefreshToken::BindErr [ip=%s [err=%v", ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.RefreshToken [ip=%s", ip)

	resp, err := core.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		commons.LogC(ctx, "[api.RefreshToken::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		if commons.IsUnauthorized(err) {
			commons.Resp401(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.RefreshToken::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, resp.User.ID)
	commons.RespOKWithMsg(c, "Token refreshed successfully", resp)
}

// Logout handles user logout
func Logout(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.Logout [user_id=%s [ip=%s", userID, ip)

	// Clear cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	commons.Log(ctx, "[api.Logout::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "Logout successful", nil)
}

// ForgotPassword initiates password reset flow
func ForgotPassword(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	var req model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.ForgotPassword::BindErr [ip=%s [err=%v", ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.ForgotPassword [email=%s [ip=%s", req.Email, ip)

	err := core.ForgotPassword(ctx, req)
	if err != nil {
		commons.LogC(ctx, "[api.ForgotPassword::KO [time_taken=%v [ip=%s [email=%s [err=%v",
			time.Since(st), ip, req.Email, err)
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ForgotPassword::OK [time_taken=%v [ip=%s [email=%s",
		time.Since(st), ip, req.Email)
	commons.RespOKWithMsg(c, "If the email exists, a password reset link has been sent", nil)
}

// ResetPassword resets password with token
func ResetPassword(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.ResetPassword::BindErr [ip=%s [err=%v", ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.ResetPassword [user_id=%s [ip=%s", req.UserID, ip)

	err := core.ResetPassword(ctx, req)
	if err != nil {
		commons.LogC(ctx, "[api.ResetPassword::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, req.UserID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ResetPassword::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, req.UserID)
	commons.RespOKWithMsg(c, "Password reset successfully", nil)
}

// ChangePassword changes user password (authenticated)
func ChangePassword(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.ChangePassword::BindErr [user_id=%s [ip=%s [err=%v", userID, ip, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.ChangePassword [user_id=%s [ip=%s", userID, ip)

	err := core.ChangePassword(ctx, userID, req)
	if err != nil {
		commons.LogC(ctx, "[api.ChangePassword::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		if commons.IsUnauthorized(err) {
			commons.Resp401(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.ChangePassword::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "Password changed successfully", nil)
}
