package controllers

import (
	"SeproWAF/models"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
)

// UIController handles UI page rendering
type UIController struct {
	web.Controller
}

// Auth pages
func (c *UIController) Login() {
	c.Data["Title"] = "Login"
	c.Layout = "layout.tpl"
	c.TplName = "auth/login.tpl"
}

func (c *UIController) Register() {
	c.Data["Title"] = "Register"
	c.Layout = "layout.tpl"
	c.TplName = "auth/register.tpl"
}

// Dashboard page
func (c *UIController) Dashboard() {
	// Try to get user from Authorization header first
	user := c.GetUserFromAuthHeader()

	// If not found in header, try cookie
	if user == nil {
		// Try localStorage via JavaScript check
		c.Data["Title"] = "Dashboard"
		c.Data["CheckAuth"] = true // Flag to trigger client-side auth check
		c.Layout = "layout.tpl"
		c.TplName = "dashboard/index.tpl"
		return
	}

	// User is authenticated
	c.Data["Title"] = "Dashboard"
	c.Data["Username"] = user.Username
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()
	c.Layout = "layout.tpl"
	c.TplName = "dashboard/index.tpl"
}

// User profile
func (c *UIController) UserProfile() {
	// Get authenticated user with the more complete method
	user := c.GetUserFromJWT()
	if user == nil {
		c.Ctx.Redirect(302, "/auth/login")
		return
	}

	// Successful authentication
	c.Data["Title"] = "User Profile"
	c.Data["Username"] = user.Username
	c.Data["Email"] = user.Email
	c.Data["Role"] = user.Role
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()

	c.Layout = "layout.tpl"
	c.TplName = "user/profile.tpl"
}

// User management (admin only)
func (c *UIController) UserList() {
	user := c.GetUserFromJWT()
	if user == nil {
		c.Redirect("/auth/login", 302)
		return
	}

	if !user.IsAdmin() {
		c.Redirect("/dashboard", 302)
		return
	}

	c.Data["Title"] = "User Management"
	c.Data["Username"] = user.Username
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()
	c.Layout = "layout.tpl"
	c.TplName = "user/list.tpl"
}

// SITE MANAGEMENT UI METHODS (Integrated from SiteUIController)

// SiteList renders the site list page
func (c *UIController) SiteList() {
	// Get authenticated user
	user := c.GetUserFromJWT()
	if user == nil {
		c.Redirect("/auth/login", 302)
		return
	}

	c.Data["Title"] = "Protected Sites"
	c.Data["Username"] = user.Username
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()
	c.Layout = "layout.tpl"
	c.TplName = "site/list.tpl"
}

// SiteDetail renders the site detail page
func (c *UIController) SiteDetail() {
	// Get site ID from URL parameter
	siteIDStr := c.Ctx.Input.Param(":id")
	siteID, err := strconv.Atoi(siteIDStr)
	if err != nil {
		c.CustomAbort(400, "Invalid site ID")
		return
	}

	// Get site from database
	site, err := models.GetSiteByID(siteID)
	if err != nil {
		c.CustomAbort(404, "Site not found")
		return
	}

	// Check if user has permission to view this site
	// Your existing permission check code here...

	// Get the active tab from query parameter, default to "overview"
	activeTab := c.GetString("tab", "overview")

	// Add data to template
	c.Data["PageTitle"] = site.Name
	c.Data["Site"] = site
	c.Data["ActiveTab"] = activeTab // Set the active tab
	c.Layout = "layout.tpl"
	c.TplName = "site/detail.tpl"
}

// SiteCreate renders the site creation page
func (c *UIController) SiteCreate() {
	// Get authenticated user
	user := c.GetUserFromJWT()
	if user == nil {
		c.Redirect("/auth/login", 302)
		return
	}

	c.Data["Title"] = "Add New Site"
	c.Data["Username"] = user.Username
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()
	c.Layout = "layout.tpl"
	c.TplName = "site/create.tpl"
}

// SiteEdit renders the site edit page
func (c *UIController) SiteEdit() {
	// Get authenticated user
	user := c.GetUserFromJWT()
	if user == nil {
		c.Redirect("/auth/login", 302)
		return
	}

	// Get site ID from URL parameter
	siteID, err := strconv.Atoi(c.Ctx.Input.Param(":id"))
	if err != nil {
		c.Abort("400")
		return
	}

	// Get the site
	site := models.Site{ID: siteID}
	o := orm.NewOrm()
	err = o.Read(&site)
	if err != nil {
		c.Abort("404")
		return
	}

	// Check if user has permission to edit the site
	if !site.CanUserManageSite(user.ID, user.Role) {
		c.Redirect("/dashboard", 302)
		return
	}

	c.Data["Title"] = "Edit Site - " + site.Name
	c.Data["Username"] = user.Username
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()
	c.Data["Site"] = site
	c.Layout = "layout.tpl"
	c.TplName = "site/edit.tpl"
}

// CertificateList renders the certificate list page
func (c *UIController) CertificateList() {
	// Get authenticated user
	user := c.GetUserFromJWT()
	if user == nil {
		c.Redirect("/auth/login", 302)
		return
	}

	c.Data["Title"] = "SSL Certificates"
	c.Data["Username"] = user.Username
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()
	c.Layout = "layout.tpl"
	c.TplName = "certificate/list.tpl"
}

// CertificateUpload renders the certificate upload page
func (c *UIController) CertificateUpload() {
	// Get authenticated user
	user := c.GetUserFromJWT()
	if user == nil {
		c.Redirect("/auth/login", 302)
		return
	}

	c.Data["Title"] = "Upload Certificate"
	c.Data["Username"] = user.Username
	c.Data["IsAuthenticated"] = true
	c.Data["IsAdmin"] = user.IsAdmin()
	c.Layout = "layout.tpl"
	c.TplName = "certificate/upload.tpl"
}

// WAFRuleList shows the WAF rules for a site
func (c *UIController) WAFRuleList() {
	// Get site ID from URL parameter
	siteIDStr := c.Ctx.Input.Param(":id")
	siteID, err := strconv.Atoi(siteIDStr)
	if err != nil {
		c.CustomAbort(400, "Invalid site ID")
		return
	}

	// Get site information
	site, err := models.GetSiteByID(siteID)
	if err != nil {
		c.CustomAbort(404, "Site not found")
		return
	}

	c.Data["Site"] = site
	c.Data["SiteID"] = siteID
	c.Data["PageTitle"] = "WAF Rules - " + site.Domain
	c.Layout = "layout.tpl"
	c.TplName = "waf/rule_list.tpl"
}

// WAFRuleCreate shows the form to create a new WAF rule
func (c *UIController) WAFRuleCreate() {
	// Get site ID from URL parameter
	siteIDStr := c.Ctx.Input.Param(":id")
	siteID, err := strconv.Atoi(siteIDStr)
	if err != nil {
		c.CustomAbort(400, "Invalid site ID")
		return
	}

	// Get site information
	site, err := models.GetSiteByID(siteID)
	if err != nil {
		c.CustomAbort(404, "Site not found")
		return
	}

	c.Data["Site"] = site
	c.Data["SiteID"] = siteID
	c.Data["IsEdit"] = false
	c.Data["PageTitle"] = "Create WAF Rule - " + site.Domain
	c.Layout = "layout.tpl"
	c.TplName = "waf/rule_edit.tpl"
}

// WAFRuleEdit shows the form to edit an existing WAF rule
func (c *UIController) WAFRuleEdit() {
	// Get site ID and rule ID from URL parameters
	siteIDStr := c.Ctx.Input.Param(":id")
	ruleIDStr := c.Ctx.Input.Param(":ruleId")

	siteID, err := strconv.Atoi(siteIDStr)
	if err != nil {
		c.CustomAbort(400, "Invalid site ID")
		return
	}

	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		c.CustomAbort(400, "Invalid rule ID")
		return
	}

	// Get site information
	site, err := models.GetSiteByID(siteID)
	if err != nil {
		c.CustomAbort(404, "Site not found")
		return
	}

	// Get rule information
	rule, err := models.GetWAFRuleByID(ruleID)
	if err != nil {
		c.CustomAbort(404, "Rule not found")
		return
	}

	// Verify that the rule belongs to the site
	if rule.SiteID != siteID {
		c.CustomAbort(403, "Rule does not belong to this site")
		return
	}

	c.Data["Site"] = site
	c.Data["SiteID"] = siteID
	c.Data["Rule"] = rule
	c.Data["RuleID"] = ruleID
	c.Data["IsEdit"] = true
	c.Data["PageTitle"] = "Edit WAF Rule - " + site.Domain
	c.Layout = "layout.tpl"
	c.TplName = "waf/rule_edit.tpl"
}

// Helper to get user from JWT token
func (c *UIController) GetUserFromJWT() *models.User {
	authHeader := c.Ctx.Input.Header("Authorization")
	if authHeader == "" {
		// Try to get token from cookie
		jwtCookie := c.Ctx.GetCookie("jwt_token")
		if jwtCookie == "" {
			return nil
		}
		authHeader = "Bearer " + jwtCookie
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil
	}

	tokenString := parts[1]
	secret, _ := web.AppConfig.String("JWTSecret")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	// Get user from database
	userID := int(claims["user_id"].(float64))
	o := orm.NewOrm()
	user := models.User{ID: userID}
	err = o.Read(&user)
	if err != nil {
		return nil
	}

	return &user
}

// GetUserFromAuthHeader retrieves user from Authorization header
func (c *UIController) GetUserFromAuthHeader() *models.User {
	authHeader := c.Ctx.Input.Header("Authorization")
	if authHeader == "" {
		return nil
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil
	}

	tokenString := parts[1]
	secret, _ := web.AppConfig.String("JWTSecret")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	// Get user from database
	userID := int(claims["user_id"].(float64))
	o := orm.NewOrm()
	user := models.User{ID: userID}
	err = o.Read(&user)
	if err != nil {
		return nil
	}

	return &user
}

// GetToken extracts the JWT token from request headers or cookies
func (c *UIController) GetToken() string {
	// Try Authorization header first
	authHeader := c.Ctx.Input.Header("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try cookie next
	jwtCookie := c.Ctx.GetCookie("jwt_token")
	if jwtCookie != "" {
		return jwtCookie
	}

	// Log the headers and cookies for debugging
	logs.Debug("Headers: %v", c.Ctx.Request.Header)
	logs.Debug("Cookies: %v", c.Ctx.Request.Cookies())

	// No valid token found
	return ""
}

// GetUserFromToken parses the token and retrieves the user
func (c *UIController) GetUserFromToken(tokenString string) *models.User {
	// Parse the token
	secret, _ := web.AppConfig.String("JWTSecret")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		logs.Error("Invalid token: %v", err)
		return nil
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logs.Error("Failed to get claims from token")
		return nil
	}

	// Get user ID from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		logs.Error("Failed to get user_id from claims")
		return nil
	}

	// Fetch user from database
	o := orm.NewOrm()
	user := models.User{ID: int(userID)}
	err = o.Read(&user)
	if err != nil {
		logs.Error("Failed to find user with ID %d: %v", int(userID), err)
		return nil
	}

	return &user
}
