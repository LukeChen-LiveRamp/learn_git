package web

import (
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`

	phoneRegexPattern = `^1\d{10}$|^(0\d{2,3}-?|\(0\d{2,3}\))?[1-9]\d{4,7}(-\d{1,8})?$`
	nicknamePattern   = `^.{1,30}$`
	birthdayPattern   = `^(19|20)\d\d-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`
	aboutMePattern    = `^.{1,1000}$`
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	phoneRexExp    *regexp.Regexp
	nicknameRexExp *regexp.Regexp
	birthdayRexExp *regexp.Regexp
	aboutMeRexExp  *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		phoneRexExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		nicknameRexExp: regexp.MustCompile(nicknamePattern, regexp.None),
		birthdayRexExp: regexp.MustCompile(birthdayPattern, regexp.None),
		aboutMeRexExp:  regexp.MustCompile(aboutMePattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// REST 风格
	//server.POST("/user", h.SignUp)
	//server.PUT("/user", h.SignUp)
	//server.GET("/users/:username", h.Profile)
	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", h.SignUp)
	// POST /users/login
	ug.POST("/login", h.Login)
	// POST /users/edit
	ug.POST("/edit", h.Edit)
	// GET /users/profile
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十五分钟
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Phone    string `json:"phone"`
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var editReq EditReq
	if err := ctx.Bind(&editReq); err != nil {
		return
	}
	userId, ok := h.getUserId(ctx)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 手机号码验证
	isPhone, err := h.phoneRexExp.MatchString(editReq.Phone)
	if !h.handleValidationError(ctx, isPhone, err, "非法手机号码格式") {
		return
	}
	// 昵称验证
	isNickname, err := h.nicknameRexExp.MatchString(editReq.Nickname)
	if !h.handleValidationError(ctx, isNickname, err, "昵称最多可以包含 30 个字符") {
		return
	}
	// 生日验证
	isBirthday, err := h.birthdayRexExp.MatchString(editReq.Birthday)
	if !h.handleValidationError(ctx, isBirthday, err, "非法生日格式") {
		return
	}
	// 个人简介验证
	isAboutMe, err := h.aboutMeRexExp.MatchString(editReq.AboutMe)
	if !h.handleValidationError(ctx, isAboutMe, err, "个人简介最多可以包含 1000 个字符") {
		return
	}

	err = h.svc.Edit(ctx, domain.User{
		Id:       userId,
		Phone:    editReq.Phone,
		Nickname: editReq.Nickname,
		Birthday: editReq.Birthday,
		AboutMe:  editReq.AboutMe,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "修改成功")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	type ProfileResp struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	userId, ok := h.getUserId(ctx)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	user, err := h.svc.Profile(ctx, userId)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	resp := ProfileResp{
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		Birthday: user.Birthday,
		AboutMe:  user.AboutMe,
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *UserHandler) getUserId(ctx *gin.Context) (int64, bool) {
	sess := sessions.Default(ctx)
	userId, ok := sess.Get("userId").(int64)
	return userId, ok
}

func (h *UserHandler) handleValidationError(ctx *gin.Context, valid bool, err error, errMsg string) bool {
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return false
	}
	if !valid {
		ctx.String(http.StatusOK, errMsg)
		return false
	}
	return true
}
