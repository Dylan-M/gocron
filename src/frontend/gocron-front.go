package main
import (
      "fmt"
      "time"
      "flag"
      "strings"
      "strconv"

      "net/http"
      "expvar"

      "github.com/jsirianni/gocronlib"
)


const (
      version     string = "2.2.0"
      libVersion  string = gocronlib.Version
      errorResp   string = "Internal Server Error"
      contentType string = "plain/text"
)

var ( // Flags set in main()
      port       string
      verbose    bool
      getVersion bool
      noProxy    bool
)

var ( // Metric variables NOTE: Placeholder in order to compile
    fooCount = expvar.NewInt("foo.count")
)


func main() {
      flag.BoolVar(&getVersion, "version", false, "Get the version and then exit")
      flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
      flag.BoolVar(&noProxy, "no-proxy", false, "Listens on all interfaces - Use this for debugging only!")
      flag.StringVar(&port, "p", "8080", "Listening port for the web server")
      flag.Parse()

      if getVersion == true {
            fmt.Println("gocron-front version: " + version)
            fmt.Println("gocronlib version: " + libVersion)
            return
      }

      gocronlib.CronLog("Verbose mode enabled", verbose)
      gocronlib.CronLog("gocron-front version: " + version, verbose)
      gocronlib.CronLog("gocronlib version: " + libVersion, verbose)
      gocronlib.CronLog("Starting web server on port: " + port, verbose)


      // Start the web server
      http.HandleFunc("/", cronStatus)

      if noProxy == true {
          gocronlib.CronLog("WARNING: --no-proxy passed, listening on all interfaces. This flag should only be used for debuging only." +
              "The metrics API is exposed, without authentication. Production systems should use a proxy, such as nginx. Please view the readme at https://github.com/jsirianni/gocron", verbose)
          http.ListenAndServe(":" + port, nil)
      } else {
          http.ListenAndServe("localhost:" + port, nil)
      }
}


// Validate the request and then pass to updateDatabase()
func cronStatus(resp http.ResponseWriter, req *http.Request) {
      var (
            currentTime int = int(time.Now().Unix())
            socket = strings.Split(req.RemoteAddr, ":")
            c gocronlib.Cron
            method string = ""
      )

      switch req.Method {
      case "GET":
            method = "GET"
            c.Cronname    = req.URL.Query().Get("cronname")
            c.Account     = req.URL.Query().Get("account")
            c.Email       = req.URL.Query().Get("email")
            c.Frequency   = gocronlib.StringToInt(req.URL.Query().Get("frequency"), verbose)
            c.Lastruntime = currentTime
            c.Ipaddress   = socket[0]

            // If x = 1, set c.Site to true
            x, err  := strconv.Atoi(req.URL.Query().Get("site"))
            if err == nil && x == 1 {
                  c.Site = true
            } else {
                  c.Site = false
            }

      case "POST":
            gocronlib.CronLog("POST not yet supported: " + c.Ipaddress, verbose)
            return

      default:
            // Log an error and do not respond
            gocronlib.CronLog("Incoming request from " + c.Ipaddress + " is not a GET or POST.", verbose)
            return
      }

      if validateParams(c) == true {
            if updateDatabase(c) == true {
                  returnCreated(resp)

            } else {
                  returnServerError(resp)
            }

      } else {
            returnNotFound(resp)
            gocronlib.CronLog(method + " from " + c.Ipaddress + " not valid. Dropping.", verbose)
      }
}


// Return a 201 Created
func returnCreated(resp http.ResponseWriter) {
      resp.Header().Set("Content-Type", contentType)
      resp.WriteHeader(http.StatusCreated)
}


// Return a 500 Server Error
func returnServerError(resp http.ResponseWriter) {
      resp.Header().Set("Content-Type", contentType)
      resp.WriteHeader(http.StatusInternalServerError)
      resp.Write([]byte(errorResp))
}


// Return 404 Not Found
func returnNotFound(resp http.ResponseWriter) {
      resp.WriteHeader(http.StatusNotFound)
}


func updateDatabase(c gocronlib.Cron) bool {
      var (
            query  string
            result bool

            // Convert variables once and use multiple times in the query
            frequency   string = strconv.Itoa(c.Frequency)
            lastruntime string = strconv.Itoa(c.Lastruntime)
            site        string = strconv.FormatBool(c.Site)
      )

      // Insert and update if already exist
      query = "INSERT INTO gocron " +
              "(cronname, account, email, ipaddress, frequency, lastruntime, alerted, site) " +
              "VALUES ('" +
              c.Cronname + "','" + c.Account + "','" + c.Email + "','" + c.Ipaddress + "','" +
              frequency + "','" + lastruntime + "','" + "false" + "','" + site + "') " +
              "ON CONFLICT (cronname, account) DO UPDATE " +
              "SET email = " + "'" + c.Email + "'," + "ipaddress = " + "'" + c.Ipaddress + "'," +
              "frequency = " + "'" + frequency + "'," + "lastruntime = " + "'" + lastruntime + "', " +
              "site = " + "'" + site + "';"

      // Execute query
      rows, result := gocronlib.QueryDatabase(query, verbose)
      defer rows.Close()
      if result == true {
            gocronlib.CronLog("Heartbeat from " + c.Cronname + ": " + c.Account + " \n", verbose)
            return true

      } else {
            return false
      }
}


// Function validates SQL variables
func validateParams(c gocronlib.Cron) bool {

      var valid bool = false  // Flag determines the return value

      if checkLength(c) == true {
            valid = true
      }

      if verbose == true {
            if valid == true {
                  gocronlib.CronLog("Parameters from " + c.Ipaddress + " passed validation", verbose)
                  return true

            } else {
                  gocronlib.CronLog("Parameters from " + c.Ipaddress + " failed validation!", verbose)
                  return false
            }
      }

      return valid
}


// Validate that parameters are present
// Validate that ints are not -1 (failed conversion in gocronlib StringToInt())
func checkLength(c gocronlib.Cron) bool {
      if len(c.Account) == 0 {
            return false

      } else if len(c.Cronname) == 0 {
            return false

      } else if len(c.Email) == 0 {
            return false

      } else if c.Frequency == -1 {
            return false

      } else if len(c.Ipaddress) == 0 {
            return false

      } else if c.Lastruntime == -1 {
            return false

      } else {
            return true
      }
}
