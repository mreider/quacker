Here's the requirements for a little app. 

The purpose of the app is to provide a service for people who use static blog generators like jekyll or hugo, to offer a "subscribe" button so that their readers can get an email every time there is a new blog post. Since static blog generators have no way of running server side code, this is currently impossible. But the app will make it possible. It's called quacker.

Quacker's requirements are as follows.

The personas are:

"admin" - the person hosting the quacker instance (it's open source).
"user" - the person hosting the static blog site
"subscriber" - the person wanting to subscribe to a user's blog posts

User Story 1: As an admin, I want to setup the app. I will do this by running "./quacker --setup". This will prompt me for my mailgun api key, which I will validate against mailguns api endpoint. It will also prompt me for my host name (not http needed), which I will also validate as a valid domain name format (*foo.com). Finally it will check if certbot is installed, and if it is, it will generate a certificate and ask the questions, etc. If any of this fais validation just error and exit. If these prompts validate / succeed, store values in redis.

If any of the other quacker commands are executed without successful setup, tell the user they need to run setup and exit.

User Story 2; As an admin, I want to be able to generate invitation codes for my users. I will do this on the command line of the linux host using ./quacker--generate"

User Story 3: As an admin I want to run the web app using "./quacker --run" which will use python's http server and serve via port 443 using the certificate we created.

User Story 4: As a user, when I go to the "home" (index) page - I will see a text input where I can enter an invitation code. When I click "submit" I will get a message that tells me if it is valid, or invalid. There will be simple flood control on this form which allows 1 submission every 2 seconds to prevent hacker bots.

User Story 5: As a user, if I submit a valid invitation code, I will see an HTML page with a form. Above the form are instructions for adding a txt record to the user's DNS record where the static blog is hosted. The user must create a txt record with the value "quack-quack". The form appears underneath the instructions includes:

    RSS feed URL for my static blog.
    Reply-to address for the emails I will send.

When the user submits, all of these will be validated such that the reply to is a valid email format, and the RSS feed for the blog responds with valid RSS XML with blog posts. There will also be a validation on the RSS feed's domain DNS information. It will check to see that there is a txt record in the DNS for that domain with the words "quack-quack' - There will be an error if any these validations fails

User story 6: If the validation succeeds, the user's website is recorded redis.

User story 7: Also on the home / index page, below the invitation code text input, there is a table showing user's websites that are in the sites stored in Redis. The websites are sorted alphabetically by domain name. There are only two columns in the table. The domain name of the site ("foo.com") and a button that says "JS"

If redis is empty, it just says "no sites" .

User story 8: When a user clicks on the "JS" button for a website, they see a page with a text area. Above the text area is a "copy" button to put the contents on their clipboard. In the text area is well formatted (readable) Javascript for them to put in their static website. The Javascript creates a simple text input and small button ("go") for subscribers to enter their email address. There is javascript to validate if it's a valid email address, and an error appears in the subscriber's browser if it is not. If it's valid, it will do the following:

    The email address is submitted (HTTP POST) to the admin's instance of quacker (the host name is in redis - the admin needs to set).
    The http referrer is checked against the websites in redis - If there is no match, the subscriber will see an error "there is no subscriber support for this blog's domain: {http referrer}"
    If the http referrer IS in redis then the subscriber's email is stored in redis associated with the site that the subscriber is a part. The same email cannot be added twice for the same site but can be added for different sites. Success or failure messages are returned and displayed.
    The quacker app code will have flood control here as well, so people can't spam subscriber emails. Just a single post every 3 seconds.


User story 9: admins should be able am to run './quacker --job' every 5 minutes via crontab. The job reads the list of sites from redis and gets all blog posts, per site, for the last three days. It then sends well formatted html emails to each subscriber using the subscribers stored in redis, for each site. The emails will include the blog post title as the subject line, The title (as an html link), description, and image (if there is one), in the email body. There is also an unsubscribe link (described later). This email is sent to each subscriber for each blog. We will use mailgun and the mailgun api key along with the admin's domain, both stored in redis during setup.

User story 10: When an email is sent to a subscriber for a specific user's blog entry, a sent record is stored in redis. When the job reads the sites and gets all blog posts, per site, it will NOT send an additional email if one was already sent for that blog entry. Subscribers only get one email for each blog entry, even though the job runs every 5 minutes. No spam.

User story 11: When a subscriber clicks the unsubscribe email, they will that "{email address} has been unsubscribed from {site domain}" - the subscriber will then be removed from redis for that site - make sure it's only removed from that site in case the email appears for other sites as well.

User story 12: Above the list of websites on the home (index) page, in small font, should appear the following text "If you want to remove your site from the list, just delete the txt record you added ("quack-quack") and your site will be removed automatically. This can take a few minutes.

User story 13: The job will also check the DNS records of each site, and if the txt record ("quack-quack") is missing for a site, it will remove it from redis. It will also remove all of the subscriber email addresses for that site.

User story 14: running "./quacker --job" will also clean the sent records in redis of everything that is older than four days to keep the redis db small.