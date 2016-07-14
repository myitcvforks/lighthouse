package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nwidger/lighthouse"
	"github.com/nwidger/lighthouse/bins"
	"github.com/nwidger/lighthouse/messages"
	"github.com/nwidger/lighthouse/milestones"
	"github.com/nwidger/lighthouse/projects"
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
project-specific.  These commands require the project ID to be
specified via -p, --project, the LH_PROJECT environment variable or
the config file.

The default config file is $HOME/.lh.yaml but can be overridden with
--config.

`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		account, token, email, password := viper.GetString("account"), viper.GetString("token"),
			viper.GetString("email"), viper.GetString("password")
		if len(account) == 0 {
			log.Fatal("Please specify Lighthouse account name via -a, --account, LH_ACCOUNT or config file")
		}
		var client *http.Client
		if len(token) > 0 {
			client = lighthouse.NewClient(token)
		} else if len(email) > 0 && len(password) > 0 {
			pw := password
			if strings.HasPrefix(password, "@") && len(password) > 1 {
				buf, err := ioutil.ReadFile(password[1:])
				if err != nil {
					log.Fatal(err)
				}
				pw = strings.TrimSpace(string(buf))
			}
			client = lighthouse.NewClientBasicAuth(email, pw)
		} else {
			log.Fatal("Please specify token or email & password")
		}
		service = lighthouse.NewService(account, client)
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
	RootCmd.PersistentFlags().StringP("project", "p", "", "Lighthouse project ID")
	viper.BindPFlag("account", RootCmd.PersistentFlags().Lookup("account"))
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("email", RootCmd.PersistentFlags().Lookup("email"))
	viper.BindPFlag("password", RootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("project", RootCmd.PersistentFlags().Lookup("project"))
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
	buf, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
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

func Users() (map[string]*projects.User, error) {
	projectID := Project()
	p := projects.NewService(service)
	ms, err := p.Memberships(projectID)
	if err != nil {
		return nil, err
	}
	userMap := map[string]*projects.User{}
	for _, m := range ms {
		userMap[m.User.Name] = m.User
	}
	return userMap, nil
}

func UserID(userStr string) (int, error) {
	id, err := strconv.ParseInt(userStr, 10, 64)
	if err == nil {
		return int(id), nil
	}
	us, err := Users()
	if err != nil {
		return 0, err
	}
	u, ok := us[userStr]
	if !ok {
		return 0, fmt.Errorf("no such user %q", userStr)
	}
	return u.ID, nil
}

func Milestones() (map[string]*milestones.Milestone, error) {
	projectID := Project()
	m := milestones.NewService(service, projectID)
	ms, err := m.ListAll(&milestones.ListOptions{})
	if err != nil {
		return nil, err
	}
	milestonesMap := map[string]*milestones.Milestone{}
	for _, milestone := range ms {
		milestonesMap[milestone.Title] = milestone
	}
	return milestonesMap, nil
}

func MilestoneID(milestoneStr string) (int, error) {
	id, err := strconv.ParseInt(milestoneStr, 10, 64)
	if err == nil {
		return int(id), nil
	}
	ms, err := Milestones()
	if err != nil {
		return 0, err
	}
	m, ok := ms[milestoneStr]
	if !ok {
		return 0, fmt.Errorf("no such milestone %q", milestoneStr)
	}
	return m.ID, nil
}

func Projects() (map[string]*projects.Project, error) {
	p := projects.NewService(service)
	ps, err := p.List()
	if err != nil {
		return nil, err
	}
	projectsMap := map[string]*projects.Project{}
	for _, project := range ps {
		projectsMap[project.Name] = project
	}
	return projectsMap, nil
}

func ProjectID(projectStr string) (int, error) {
	id, err := strconv.ParseInt(projectStr, 10, 64)
	if err == nil {
		return int(id), nil
	}
	ps, err := Projects()
	if err != nil {
		return 0, err
	}
	p, ok := ps[projectStr]
	if !ok {
		return 0, fmt.Errorf("no such project %q", projectStr)
	}
	return p.ID, nil
}

func Bins() (map[string]*bins.Bin, error) {
	projectID := Project()
	m := bins.NewService(service, projectID)
	ms, err := m.List()
	if err != nil {
		return nil, err
	}
	binsMap := map[string]*bins.Bin{}
	for _, bin := range ms {
		binsMap[bin.Name] = bin
	}
	return binsMap, nil
}

func BinID(binStr string) (int, error) {
	id, err := strconv.ParseInt(binStr, 10, 64)
	if err == nil {
		return int(id), nil
	}
	ms, err := Bins()
	if err != nil {
		return 0, err
	}
	m, ok := ms[binStr]
	if !ok {
		return 0, fmt.Errorf("no such bin %q", binStr)
	}
	return m.ID, nil
}

func Messages() (map[string]*messages.Message, error) {
	projectID := Project()
	m := messages.NewService(service, projectID)
	ms, err := m.List()
	if err != nil {
		return nil, err
	}
	messagesMap := map[string]*messages.Message{}
	for _, message := range ms {
		messagesMap[message.Title] = message
	}
	return messagesMap, nil
}

func MessageID(messageStr string) (int, error) {
	id, err := strconv.ParseInt(messageStr, 10, 64)
	if err == nil {
		return int(id), nil
	}
	ms, err := Messages()
	if err != nil {
		return 0, err
	}
	m, ok := ms[messageStr]
	if !ok {
		return 0, fmt.Errorf("no such message %q", messageStr)
	}
	return m.ID, nil
}
