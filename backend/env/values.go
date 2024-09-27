package env

type values struct {
	appMode            string
	appPort            string
	corsAllowedOrigins []string
	authSecretKey      string
	authCookieDomain   string
	dbUser             string
	dbPass             string
	dbHost             string
	dbPort             string
	dbName             string
}

type Values interface {
	AppMode() string
	AppPort() string
	CorsAllowedOrigins() []string
	AuthSecretKey() string
	AuthCookieDomain() string
	DBUser() string
	DBPass() string
	DBHost() string
	DBPort() string
	DBName() string
}

func (v *values) AppMode() string {
	return v.appMode
}

func (v *values) AppPort() string {
	return v.appPort
}

func (v *values) CorsAllowedOrigins() []string {
	return v.corsAllowedOrigins
}

func (v *values) AuthSecretKey() string {
	return v.authSecretKey
}

func (v *values) AuthCookieDomain() string {
	return v.authCookieDomain
}

func (v *values) DBUser() string {
	return v.dbUser
}

func (v *values) DBPass() string {
	return v.dbPass
}

func (v *values) DBHost() string {
	return v.dbHost
}

func (v *values) DBPort() string {
	return v.dbPort
}

func (v *values) DBName() string {
	return v.dbName
}
