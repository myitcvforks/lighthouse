package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nwidger/jsoncolor"
	"github.com/nwidger/lighthouse"
	"github.com/nwidger/lighthouse/milestones"
	"github.com/nwidger/lighthouse/projects"
	"github.com/nwidger/lighthouse/users"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	service *lighthouse.Service
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lh",
	Short: "lh provides CLI access to the Lighthouse API http://help.lighthouseapp.com/kb/api",
	Long: `lh provides CLI access to the Lighthouse API http://help.lighthouseapp.com/kb/api

Please specify your Lighthouse account name via -a, --account, the
LH_ACCOUNT environment variable or the config file.  If your
Lighthouse URL is 'https://your-account-name.lighthouseapp.com' then
your account name is 'your-account-name'.

Lighthouse requires a valid API token or email/password to
authenticate API requests.  Please specify a Lighthouse API token via
-t, --token, the LH_TOKEN environment variable or the config file.  If
you'd prefer to authenticate with an email/password, please specify it
via -e, --email, the LH_EMAIL environment variable, -p, --password,
the LH_PASSWORD environment variable or the config file.  If the
specified password has the form '@FILE', the password is instead read
from FILE.

Many subcommands work on resources that are Lighthouse
project-specific.  These commands require the project ID or name to be
specified via -p, --project, the LH_PROJECT environment variable or
the config file.

On Unix systems, the default config file is $HOME/.lh.yaml.  On
Windows systems, the default config file is
%HOMEDRIVE%\%HOMEPATH%\.lh.yaml, falling back to
%USERPROFILE%\.lh.yaml if necessary.  On all systems, the default can
be overridden with --config.

`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		account, token, email, password, interval, burstSize := viper.GetString("account"), viper.GetString("token"),
			viper.GetString("email"), viper.GetString("password"),
			viper.GetDuration("rate-limit-interval"), viper.GetInt("rate-limit-burst-size")
		if len(account) == 0 {
			FatalUsage(cmd, "Please specify Lighthouse account name via -a, --account, LH_ACCOUNT or config file")
		}
		lt := &lighthouse.Transport{
			TokenAsBasicAuth: true,
		}
		client := &http.Client{
			Transport: lt,
		}
		if len(token) > 0 {
			lt.Token = token
		} else if len(email) > 0 && len(password) > 0 {
			pw := password
			if strings.HasPrefix(password, "@") && len(password) > 1 {
				buf, err := ioutil.ReadFile(password[1:])
				if err != nil {
					FatalUsage(cmd, err)
				}
				pw = strings.TrimSpace(string(buf))
			}
			lt.Email = email
			lt.Password = pw
		} else {
			FatalUsage(cmd, "Please specify token or email & password")
		}
		if interval != time.Duration(0) {
			lt.RateLimitInterval = interval
			lt.RateLimitBurstSize = burstSize
		}
		service = lighthouse.NewService(account, client)
		service.RateLimitRetryRequests = true
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lh.yaml)")
	RootCmd.PersistentFlags().StringP("account", "a", "", "Lighthouse account name")
	RootCmd.PersistentFlags().StringP("token", "t", "", "Lighthouse API token")
	RootCmd.PersistentFlags().String("email", "", "Lighthouse email (cannot be used with --token)")
	RootCmd.PersistentFlags().String("password", "", "Lighthouse password (cannot be used with --token)")
	RootCmd.PersistentFlags().StringP("project", "p", "", "Lighthouse project ID or name")
	RootCmd.PersistentFlags().BoolP("monochrome", "M", false, "Monochrome (don't colorize JSON)")
	RootCmd.PersistentFlags().DurationP("rate-limit-interval", "r", lighthouse.DefaultRateLimitInterval, "Interval used to rate limit API requests (use 0 to disable rate limiting)")
	RootCmd.PersistentFlags().IntP("rate-limit-burst-size", "b", lighthouse.DefaultRateLimitBurstSize, "Burst size used to rate limit API requests (must be used with --rate-limit-interval)")
	viper.BindPFlag("account", RootCmd.PersistentFlags().Lookup("account"))
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("email", RootCmd.PersistentFlags().Lookup("email"))
	viper.BindPFlag("password", RootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("project", RootCmd.PersistentFlags().Lookup("project"))
	viper.BindPFlag("monochrome", RootCmd.PersistentFlags().Lookup("monochrome"))
	viper.BindPFlag("rate-limit-interval", RootCmd.PersistentFlags().Lookup("rate-limit-interval"))
	viper.BindPFlag("rate-limit-burst-size", RootCmd.PersistentFlags().Lookup("rate-limit-burst-size"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".lh")   // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.SetEnvPrefix("lh")     // will be uppercased automatically
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func JSON(v interface{}) {
	marshalIndent := jsoncolor.MarshalIndent
	if viper.GetBool("monochrome") {
		marshalIndent = json.MarshalIndent
	}
	buf, err := marshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
}

func Account() string {
	account := viper.GetString("account")
	if len(account) == 0 {
		log.Fatal("Please specify account name via -a, --account, LH_ACCOUNT or config file")
	}
	return account
}

func Project() int {
	projectStr := viper.GetString("project")
	if len(projectStr) == 0 {
		log.Fatal("Please specify project ID via -p, --project, LH_PROJECT or config file")
	}
	projectID, err := ProjectID(projectStr)
	if err != nil {
		log.Fatal(err)
	}
	return projectID
}

func UserID(userStr string) (int, error) {
	s := users.NewService(service)
	u, err := s.Get(userStr)
	if err != nil {
		return 0, err
	}
	return u.ID, nil
}

func MilestoneID(milestoneStr string) (int, error) {
	projectID := Project()
	s := milestones.NewService(service, projectID)
	m, err := s.Get(milestoneStr)
	if err != nil {
		return 0, err
	}
	return m.ID, nil
}

func ProjectID(projectStr string) (int, error) {
	s := projects.NewService(service)
	p, err := s.Get(projectStr)
	if err != nil {
		return 0, err
	}
	return p.ID, nil
}

func FatalUsage(cmd *cobra.Command, v ...interface{}) {
	fmt.Println(v...)
	fmt.Println()
	cmd.Usage()
	os.Exit(1)
}
