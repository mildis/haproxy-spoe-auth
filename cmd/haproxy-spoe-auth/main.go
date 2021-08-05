package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/criteo/haproxy-spoe-auth/internal/agent"
	"github.com/criteo/haproxy-spoe-auth/internal/auth"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func LogLevelFromLogString(level string) logrus.Level {
	switch level {
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "The path to the configuration file")
	flag.Parse()

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config") // name of config file (without extension)
		viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath(".")      // optionally look for config in the working directory
	}
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		logrus.Panic(fmt.Errorf("fatal error config file: %w", err))
	}

	logrus.SetLevel(LogLevelFromLogString(viper.GetString("server.log_level")))

	ldapAuthentifier := auth.NewLDAPAuthenticator(auth.LDAPConnectionDetails{
		Hostname:   viper.GetString("ldap.hostname"),
		Port:       viper.GetInt("ldap.port"),
		UserDN:     viper.GetString("ldap.user_dn"),
		Password:   viper.GetString("ldap.password"),
		BaseDN:     viper.GetString("ldap.base_dn"),
		UserFilter: viper.GetString("ldap.user_filter"),
	})

	oidcAuthenticator := auth.NewOIDCAuthenticator(auth.OIDCAuthenticatorOptions{
		OAuth2AuthenticatorOptions: auth.OAuth2AuthenticatorOptions{
			ClientID:        viper.GetString("oidc.client_id"),
			ClientSecret:    viper.GetString("oidc.client_secret"),
			RedirectURL:     viper.GetString("oidc.redirect_url"),
			CallbackAddr:    viper.GetString("oidc.callback_addr"),
			CookieName:      viper.GetString("oidc.cookie_name"),
			CookieDomain:    viper.GetString("oidc.cookie_domain"),
			CookieSecure:    viper.GetBool("oidc.cookie_secure"),
			CookieTTL:       viper.GetDuration("oidc.cookie_ttl_seconds") * time.Second,
			SignatureSecret: viper.GetString("oidc.signature_secret"),
			Scopes:          viper.GetStringSlice("oidc.scopes"),
		},
		ProviderURL:      viper.GetString("oidc.provider_url"),
		EncryptionSecret: viper.GetString("oidc.encryption_secret"),
	})

	agent.StartAgent(viper.GetString("server.addr"), map[string]auth.Authenticator{
		"try-auth-ldap": ldapAuthentifier,
		"try-auth-oidc": oidcAuthenticator,
	})
}
