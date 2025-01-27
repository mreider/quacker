package main

import (
	"net/http"
	"fmt"
	"sort"
	"strconv"
	"time"
)

func aboutPage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("quacker_user")
	userLoggedIn := (err == nil)
	loggedInUser := ""
	if userLoggedIn {
		loggedInUser = cookie.Value
	}

	w.Write([]byte(`<!DOCTYPE html>
	<html>
	<head>
		<link rel="stylesheet" href="/assets/css/bootstrap.min.css">
		<script src="/assets/js/bootstrap.bundle.min.js"></script>
	</head>
	<body>
		<div class="container mt-5">
			` + renderMenu(r) + `
			<div class="card">
				<div class="card-body">
	`))

	if userLoggedIn {
		// Fetch activity data from Redis
		sites, _ := rdb.Keys(ctx, "user_sites:"+loggedInUser+":*").Result()
		var activity []struct {
			Date        time.Time
			BlogTitle   string
			Subscribers int
		}

		for _, siteKey := range sites {
			posts, _ := rdb.ZRevRangeWithScores(ctx, "activity:"+siteKey, 0, -1).Result()
			for _, post := range posts {
				subscribersStr := rdb.Get(ctx, "subscribers:"+siteKey+":"+post.Member.(string)).Val()
				subscribers, _ := strconv.Atoi(subscribersStr)
				activity = append(activity, struct {
					Date        time.Time
					BlogTitle   string
					Subscribers int
				}{
					Date:        time.Unix(int64(post.Score), 0),
					BlogTitle:   post.Member.(string),
					Subscribers: subscribers,
				})
			}
		}

		// Sort by date descending
		sort.Slice(activity, func(i, j int) bool {
			return activity[i].Date.After(activity[j].Date)
		})

		// Limit to the last 10 posts
		if len(activity) > 10 {
			activity = activity[:10]
		}

		if len(activity) == 0 {
			w.Write([]byte(`<h1>Latest Activity</h1><p>No activity.</p>`))
		} else {
			w.Write([]byte(`<h1>Latest Activity</h1><table class="table table-striped"><thead><tr><th>Date Sent</th><th>Blog Title</th><th>Subscribers</th></tr></thead><tbody>`))
			for _, entry := range activity {
				w.Write([]byte(fmt.Sprintf(`<tr><td>%s</td><td><a href="%s">%s</a></td><td>%d</td></tr>`, entry.Date.Format("2006-01-02 15:04:05"), entry.BlogTitle, entry.BlogTitle, entry.Subscribers)))
			}
			w.Write([]byte(`</tbody></table>`))
		}
	} else {
		w.Write([]byte(`
				<h1>Welcome to Quacker</h1>
				<p>Quacker is a service designed to add subscription capabilities to static websites generated by tools like Jekyll or Hugo.</p>
				<a href="/login/github" class="btn btn-primary mt-3">Login with GitHub</a>
		`))
	}

	w.Write([]byte(`
				</div>
			</div>
			<footer class="mt-4 text-center">
				<small>&copy; 2025 Matthew Reider - <a href="https://github.com/mreider">GitHub</a></small>
			</footer>
		</div>
	</body>
	</html>`))
}