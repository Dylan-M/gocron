package libgocron

import (
    "errors"
    "net/url"
)

// Version holds the current version of gocron
const Version string = "5.0"

var config Config
var verbose bool

// Cron defines a cronjob
type Cron struct {
      Cronname    string `json:cronname`  // Name of the cronjob
      Account     string `json:account`   // Account the job belongs to
      Email       string `json:email`     // Address to send alerts to
      Frequency   int    `json:frequency` // How often a job should check in
      Site        bool   `json:site`      // Set true if service is a site (Example: Network gateway)
      Ipaddress   string   // Source IP address
      Lastruntime int      // Unix timestamp
      Alerted     bool     // set to true if an alert has already been thrown
}

// Config defines a global configuration used at runtime
type Config struct {
      Dbfqdn       string `yaml:"dbfqdn"`
      Dbport       string `yaml:"dbport"`
      Dbuser       string `yaml:"dbuser"`
      Dbpass       string `yaml:"dbpass"`
      Dbdatabase   string `yaml:"dbdatabase"`
      Interval     int    `yaml:"interval"`
      SlackHookURL string `yaml:"slackhookurl"`
      SlackChannel string `yaml:"slackchannel"`
      PreferSlack  bool   `yaml:"preferslack"`
}

// Validate checks if config parameters are valid
func (c Config) Validate() error {
    message := "Errors found in the configuration:\n"
    m := ""

    if len(c.Dbdatabase) == 0 {
        m = m + "dbdatabase is length 0\n"
    }

    if len(c.Dbfqdn) == 0 {
        m = m + "dbfqdn is length 0\n"
    }

    if len(c.Dbpass) == 0 {
        m = m + "dbpass is length 0\n"
    }

    if len(c.Dbport) == 0 {
        m = m + "dbport is length 0\n"
    }

    if len(c.Dbuser) == 0 {
        m = m + "dbuser is length 0\n"
    }

    if c.Interval < 1 {
        m = m + "interval is less than 1\n"
    }

    //if len(c.PreferSlack) == 0 {
    //}

    if len(c.SlackChannel) == 0 {
        m = m + "slack_channel is length 0\n"
    }

    if len(c.SlackHookURL) == 0 {
        m = m + "slack_hook_url is length 0\n"
    } else {
        _, err := url.Parse(c.SlackHookURL)
        if err != nil {
            m = m + "slack_hook_url: "
            m = m + err.Error()
        }
    }

    if len(m) > 0 {
        return errors.New(message + m)
    }
    return nil
}
