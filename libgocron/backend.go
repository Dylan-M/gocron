package libgocron
import (
	"fmt"
	"time"

	"github.com/jsirianni/slacklib/slacklib"
)


// Timer calls checkCronStatus() on a set interval
func Timer(verbose bool) {
	for {
		time.Sleep((time.Duration(config.Interval) * time.Second))
		CronLog("Checking for missed jobs.", verbose)
		cronStatus(verbose)
	}
}


func cronStatus(verbose bool) {
	checkMissedJobs(missedJobs, verbose)
	checkRevivedJobs(revivedJobs, verbose)
}


func checkMissedJobs(query string, verbose bool) {
	rows, status := QueryDatabase(query, verbose)
	defer rows.Close()
	if status == false {
		CronLog("Failed to perform query: "+query, verbose)
		return
	}

	for rows.Next() {
		var cron Cron
		rows.Scan(&cron.Cronname,
			&cron.Account,
			&cron.Email,
			&cron.Ipaddress,
			&cron.Frequency,
			&cron.Lastruntime,
			&cron.Alerted,
			&cron.Site)


		if cron.Alerted != true {
			subject := cron.Cronname + ": " + cron.Account + " failed to check in" + "\n"
			message := "The cronjob " + cron.Cronname + " for account " + cron.Account + " has not checked in on time"

			// Only update database if alert sent successful
			if alert(cron, subject, message, verbose) == true {
				query = "UPDATE gocron SET alerted = true " +
					"WHERE cronname = '" + cron.Cronname + "' AND account = '" + cron.Account + "';"

				rows, result := QueryDatabase(query, verbose)
				defer rows.Close()
				if result == false {
					CronLog("Failed to update row for " + cron.Cronname, verbose)
				}
			}

		} else {
			CronLog("Alert for "+cron.Cronname+": "+cron.Account+
				" has been supressed. Already alerted", verbose)
		}
	}
}


func checkRevivedJobs(query string, verbose bool) {
	rows, status := QueryDatabase(query, verbose)
	defer rows.Close()
	if status == false {
		CronLog("Failed to perform query: "+query, verbose)
		return
	}

	for rows.Next() {
		var cron Cron
		rows.Scan(&cron.Cronname,
			&cron.Account,
			&cron.Email,
			&cron.Ipaddress,
			&cron.Frequency,
			&cron.Lastruntime,
			&cron.Alerted,
			&cron.Site)

		query = "UPDATE gocron SET alerted = false " +
			"WHERE cronname = '" + cron.Cronname + "' AND account = '" + cron.Account + "';"

		rows, result := QueryDatabase(query, verbose)
		defer rows.Close()
		if result == false {
			CronLog("Failed to update row for " + cron.Cronname, verbose)

		}

		subject := cron.Cronname + ": " + cron.Account + " is back online" + "\n"
		message := "The cronjob " + cron.Cronname + " for account " + cron.Account + " is back online"
		alert(cron, subject, message, verbose)
	}
}


func GetSummary(verbose bool) {
	var message string = "gocron summary - missed jobs:\n"

	rows, status := QueryDatabase(missedJobs, verbose)
	defer rows.Close()
	if status == false {
		CronLog("Failed to perform query while attempting to build a summary: " + missedJobs, verbose)
		return
	}

	for rows.Next() {
		var cron Cron
		rows.Scan(&cron.Cronname,
			&cron.Account,
			&cron.Email,
			&cron.Ipaddress,
			&cron.Frequency,
			&cron.Lastruntime,
			&cron.Alerted,
			&cron.Site)

		message = message + "Name: " + cron.Cronname  + "| Account: " + cron.Account + "\n"
	}


	// Send slack alert and pass dummy cron object
	if verbose == true && slackAlert("gocron alert summary", message) == true {
		CronLog(message, verbose)
		return

	} else if verbose == false {
		fmt.Println(message)

	} else {
		CronLog("GOCRON: Failed to build alert summary.", verbose)
	}

}


func alert(cron Cron, subject string, message string, verbose bool) bool {

    // Immediately log the alert
    CronLog(subject, verbose)

    var result bool = false
	if slackAlert(subject, message) == true {
		result = true
	}

	// NOTE: future alert methods will go here. Removed SMTP due to complexity

    if result == true {
        CronLog("gocron success: alert for " + cron.Cronname + " sent", verbose)
        return true
    } else {
        CronLog("gocron fail: alert for " + cron.Cronname, verbose)
        return false
    }
}


func slackAlert(subject string, message string) bool {
    var slackmessage slacklib.SlackPost
    slackmessage.Channel = config.SlackChannel
    slackmessage.Text = message
    return slacklib.BasicMessage(slackmessage, config.SlackHookUrl)
}
