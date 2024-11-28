import os
import requests
from uuid import uuid4
from urllib.parse import urlparse
from dateutil import parser
from pytz import utc
import datetime
from dotenv import load_dotenv
from email_validator import validate_email, EmailNotValidError
from flask import Flask, render_template, request, redirect, url_for, flash, jsonify, session
from flask_sqlalchemy import SQLAlchemy
from flask_migrate import Migrate
from flask_wtf.csrf import CSRFProtect
from flask import got_request_exception
from apscheduler.schedulers.background import BackgroundScheduler
from cryptography.fernet import Fernet
from flask_cors import CORS
from bs4 import BeautifulSoup
from sqlalchemy.exc import IntegrityError
import textwrap

import logging


# Scheduler
scheduler = BackgroundScheduler()

# Flask app setup
app = Flask(__name__, static_folder='static', template_folder='templates')
load_dotenv()
SHOW_SHAME_MESSAGE = os.getenv('SHOW_SHAME_MESSAGE', 'true').lower() == 'true'
SHOW_WARNING_NOTE = os.getenv('SHOW_WARNING_NOTE', 'true').lower() == 'true'
REQUIRE_INVITE_CODES = os.getenv('REQUIRE_INVITE_CODES', 'true').lower() == 'true'
SECRET_KEY = os.getenv('SECRET_KEY')
FERNET_KEY = os.getenv('FERNET_KEY')
ADMIN_MAILGUN_API_KEY = os.getenv('ADMIN_MAILGUN_API_KEY')
MAILGUN_DOMAIN = os.getenv('MAILGUN_DOMAIN')
APP_HOST_URL = os.getenv('APP_HOST_URL', 'http://localhost:5000/')
BUY_ME_A_COFFEE_LINK = os.getenv('BUY_ME_A_COFFEE_LINK', 'https://buymeacoffee.com/mreider')
cipher = Fernet(FERNET_KEY)
basedir = os.path.abspath(os.path.dirname(__file__))
app.config['SQLALCHEMY_DATABASE_URI'] = f"sqlite:///{os.path.join(basedir, 'quacker.db')}"
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.config['SECRET_KEY'] = SECRET_KEY
db = SQLAlchemy(app)
migrate = Migrate(app, db)
csrf = CSRFProtect(app)
csrf.init_app(app)

# Logging setup
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("QuackerApp")

CORS(app)

# Models
class Account(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    newsletter_email = db.Column(db.String(120), nullable=False)
    _mailgun_api_key = db.Column(db.String(255), nullable=True)
    mailgun_domain = db.Column(db.String(255), nullable=False)  # New field
    delete_token = db.Column(db.String(255), nullable=True)
    created_at = db.Column(db.DateTime, default=db.func.now())

    @property
    def mailgun_api_key(self):
        return cipher.decrypt(self._mailgun_api_key.encode()).decode()

    @mailgun_api_key.setter
    def mailgun_api_key(self, value):
        self._mailgun_api_key = cipher.encrypt(value.encode()).decode()

class InviteCode(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    code = db.Column(db.String(50), unique=True, nullable=False)
    created_at = db.Column(db.DateTime, default=db.func.now())
    used = db.Column(db.Boolean, default=False)
    expires_at = db.Column(db.DateTime, nullable=False)

class Blog(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    account_id = db.Column(db.Integer, db.ForeignKey('account.id'), nullable=False)
    rss_feed = db.Column(db.String(255), unique=True, nullable=False)
    home_url = db.Column(db.String(255), nullable=False)
    created_at = db.Column(db.DateTime, default=db.func.now())
    last_email_sent = db.Column(db.DateTime, nullable=True)
    donated_date = db.Column(db.DateTime, nullable=True)

class LogEntry(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    event_type = db.Column(db.String(50), nullable=False)  # e.g., "newsletter_sent", "blog_added"
    message = db.Column(db.String(255), nullable=False)  # Readable message
    timestamp = db.Column(db.DateTime, default=db.func.now())  # When the event occurred


class SentArticle(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    blog_id = db.Column(db.Integer, db.ForeignKey('blog.id'), nullable=False)
    article_link = db.Column(db.String(255), nullable=False, unique=True)
    date_sent = db.Column(db.Date, nullable=False)


class Subscriber(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    blog_id = db.Column(db.Integer, db.ForeignKey('blog.id'), nullable=False)
    email = db.Column(db.String(120), nullable=False)
    created_at = db.Column(db.DateTime, default=db.func.now())


def setup_database():
    with app.app_context():
        db.create_all()
        logger.info("Tables checked and created if necessary.")

# Routes
@app.context_processor
def inject_settings():
    return {
        "buy_me_a_coffee_link": BUY_ME_A_COFFEE_LINK,
        "show_warning_note": SHOW_WARNING_NOTE,
    }

@app.route('/')
def index():
    blogs = Blog.query.order_by(Blog.home_url.asc()).all()
    data = []

    for blog in blogs:
        subscriber_count = Subscriber.query.filter_by(blog_id=blog.id).count()

        # Collect donation data only if SHOW_SHAME_MESSAGE is enabled
        donated_date = "✅" if SHOW_SHAME_MESSAGE and blog.donated_date else "❌" if SHOW_SHAME_MESSAGE else None

        # Collect blog data
        data.append({
            "id": blog.id,
            "blog_url": blog.home_url,
            "subscribers": subscriber_count,
            "donated": donated_date,  # Optional based on feature flag
        })

    return render_template('index.html',blogs=data,show_shame_message=SHOW_SHAME_MESSAGE,require_invite_codes=REQUIRE_INVITE_CODES,)

@app.route('/enable-with-code', methods=['POST'])
def enable_with_code():
    invite_code = request.form.get('invite_code')

    if not invite_code:
        flash("Invite code is required.", "danger")
        return redirect(url_for('index'))

    # Validate invite code
    code = InviteCode.query.filter_by(code=invite_code, used=False).first()
    if not code or code.expires_at < datetime.datetime.now():
        flash("Invalid or expired invite code.", "danger")
        return redirect(url_for('index'))

    # Mark the code as used
    code.used = True
    db.session.commit()

    # Store invite code validation in session
    session['invite_code_valid'] = True

    flash("Invite code accepted - fill out the form to enable your blog.", "success")
    return redirect(url_for('enable_newsletters'))

@app.route('/enable-newsletters', methods=['GET', 'POST'])
def enable_newsletters():
    if REQUIRE_INVITE_CODES:
        # Check if invite code has been validated
        if not session.get('invite_code_valid'):
            flash("You must enter a valid invite code to access this page.", "danger")
            return redirect(url_for('index'))

    if request.method == 'POST':
        newsletter_email = request.form.get('newsletter_email')
        mailgun_api_key = request.form.get('mailgun_api_key')
        mailgun_domain = request.form.get('mailgun_domain')
        rss_feed = request.form.get('rss_feed')
        home_url = request.form.get('home_url')

        try:
            valid_newsletter_email = validate_email(newsletter_email).email
        except EmailNotValidError as e:
            flash(str(e), 'danger')
            return redirect(url_for('enable_newsletters'))

        account = Account(
            newsletter_email=valid_newsletter_email,
            mailgun_api_key=mailgun_api_key,
            mailgun_domain=mailgun_domain
        )
        db.session.add(account)
        db.session.commit()

        blog = Blog(account_id=account.id, rss_feed=rss_feed, home_url=home_url)
        db.session.add(blog)
        db.session.commit()
        log_event("blog_added", f"Blog added for {home_url}.")

        flash('Newsletter enabled successfully! Now generate your subscription form.', 'success')
        return redirect(url_for('show_html_code', blog_id=blog.id))

    return render_template('enable_newsletters.html')

@app.route('/unsubscribe', methods=['GET'])
def unsubscribe():
    email = request.args.get('email')
    blog_id = request.args.get('blog_id')

    if not email or not blog_id:
        return "Invalid unsubscribe request.", 400

    # Query the blog using the provided blog_id
    blog = Blog.query.get(blog_id)
    if not blog:
        return "Blog not found.", 404

    # Query the subscriber associated with the blog_id and email
    subscriber = Subscriber.query.filter_by(blog_id=blog_id, email=email).first()
    if not subscriber:
        return "Subscriber not found.", 404

    # Delete the subscriber
    db.session.delete(subscriber)
    db.session.commit()

    # Log the unsubscription event
    log_event(
    "subscriber_removed",
    f"Subscriber {mask_email(email)} unsubscribed from {blog.home_url}."
    )

    # Render the unsubscription confirmation
    return render_template('unsubscribe_confirmation.html')


@app.route('/subscribe/<int:blog_id>', methods=['GET', 'POST'])
@csrf.exempt
def subscribe(blog_id):
    blog = Blog.query.get(blog_id)
    if not blog:
        return jsonify({'error': 'Invalid subscription URL'}), 404

    if request.method == 'GET':
        # Check if the user is already subscribed
        user_email = request.args.get('email', '')  # Optional email query param
        is_subscribed = Subscriber.query.filter_by(blog_id=blog.id, email=user_email).first() is not None
        return jsonify({'is_subscribed': is_subscribed})

    if request.method == 'POST':
        # Handle subscription logic
        data = request.get_json()
        email = data.get('email')
        if not email:
            return jsonify({'error': 'Email is required'}), 400

        referrer = request.headers.get('Referer')
        if not referrer or urlparse(referrer).netloc != urlparse(blog.home_url).netloc:
            return jsonify({'error': 'Subscriptions must be submitted from the same domain as the blog.'}), 403

        if Subscriber.query.filter_by(blog_id=blog.id, email=email).first():
            return jsonify({'message': 'Already subscribed.'}), 200

        subscriber = Subscriber(blog_id=blog.id, email=email)
        db.session.add(subscriber)
        db.session.commit()
        log_event("subscriber_added", f"Subscriber {mask_email(email)} added to {blog.home_url}.")
        return jsonify({'message': 'Subscribed successfully!'}), 200

if __name__ == "__main__":
    app.run()

import textwrap

@app.route('/show-html-code/<int:blog_id>', methods=['GET'])
def show_html_code(blog_id):
    blog = Blog.query.get(blog_id)
    if not blog:
        flash("Blog not found.", "danger")
        return redirect(url_for('index'))

    # Generate the HTML form code as a formatted multi-line string
    form_code = textwrap.dedent(f"""
        <form id="subscription-form">
            <label for="email">Subscribe to our Newsletter:</label>
            <input type="email" name="email" required placeholder="Enter your email">
            <button type="submit">Subscribe</button>
        </form>
        <p id="status-message" style="display: none;"></p>
        <script>
            const form = document.getElementById("subscription-form");
            const statusMessage = document.getElementById("status-message");
            const checkSubscriptionStatus = async (email) => {{
                const url = "{APP_HOST_URL}subscribe/{blog.id}?email=" + encodeURIComponent(email);
                try {{
                    const response = await fetch(url);
                    if (response.ok) {{
                        const data = await response.json();
                        return data.is_subscribed;
                    }}
                }} catch (error) {{
                    console.error("Error checking subscription status:", error);
                }}
                return false;
            }};
            form.addEventListener("submit", async (e) => {{
                e.preventDefault();
                const emailInput = form.querySelector('input[name="email"]');
                const button = form.querySelector("button");
                const email = emailInput.value;
                button.textContent = "Please wait...";
                button.disabled = true;
                emailInput.disabled = true;
                try {{
                    const isSubscribed = await checkSubscriptionStatus(email);
                    if (isSubscribed) {{
                        form.style.display = "none";
                        statusMessage.style.display = "block";
                        statusMessage.textContent = "Already subscribed!";
                        return;
                    }}
                    const url = "{APP_HOST_URL}subscribe/{blog.id}";
                    const response = await fetch(url, {{
                        method: "POST",
                        headers: {{ "Content-Type": "application/json" }},
                        body: JSON.stringify({{ email }})
                    }});
                    if (response.ok) {{
                        form.style.display = "none";
                        statusMessage.style.display = "block";
                        statusMessage.textContent = "Subscribed successfully!";
                    }} else {{
                        form.style.display = "none";
                        statusMessage.style.display = "block";
                        statusMessage.textContent = "Something went wrong. Please try again later.";
                    }}
                }} catch (error) {{
                    console.error("Error submitting form:", error);
                    form.style.display = "none";
                    statusMessage.style.display = "block";
                    statusMessage.textContent = "Something went wrong. Please try again later.";
                }}
            }});
        </script>
    """).strip()

    return render_template('show_html_code.html', form_code=form_code)

@app.route('/disable-newsletters', methods=['POST'])
def disable_newsletters():
    email = request.form.get('email')
    blog_id = request.form.get('blog_id')

    # Validate inputs
    if not email or not blog_id:
        flash("Invalid request. Please try again.", "danger")
        return redirect(url_for('index'))

    # Fetch the blog and associated account
    blog = Blog.query.get(blog_id)
    if not blog:
        flash("Blog not found.", "danger")
        return redirect(url_for('index'))

    account = Account.query.get(blog.account_id)
    if not account:
        flash("Account not found for this blog.", "danger")
        return redirect(url_for('index'))

    # Check if the email matches the account email
    if email.lower() != account.newsletter.lower():
        flash("If the email matches the one on file for this domain, you will receive an email with a disable confirmation code.", "info")
        return redirect(url_for('index'))

    # Generate a unique token for confirmation
    token = str(uuid4())
    delete_link = f"{request.host_url}confirm-delete-account?token={token}"
    account.delete_token = token
    db.session.commit()

    # Send the confirmation email
    send_deletion_email(email, delete_link)
    flash("If the email matches the one on file for this domain, you will receive an email with a disable confirmation code.", "info")
    return redirect(url_for('index'))


@app.route('/confirm-delete-account', methods=['GET'])
def confirm_delete_account():
    token = request.args.get('token')

    account = Account.query.filter_by(delete_token=token).first()
    if not account:
        return "Invalid or expired token.", 404

    Blog.query.filter_by(account_id=account.id).delete()
    Subscriber.query.filter(Subscriber.blog_id.in_(
        db.session.query(Blog.id).filter_by(account_id=account.id)
    )).delete()
    db.session.delete(account)
    db.session.commit()
    log_event("blog_deleted",f"Blog deleted for account {account.newsletter_email}.")
    return render_template('confirm_delete_account.html')

@app.route('/admin/logs')
def admin_logs():
    logs = LogEntry.query.order_by(LogEntry.timestamp.desc()).all()
    return render_template('admin_logs.html', logs=logs)


# Helper Functions

def log_event(event_type, message):
    """Log events with optional email masking."""
    words = message.split()
    for i, word in enumerate(words):
        if "@" in word and "." in word:  # Naive check for email format
            words[i] = mask_email(word)
    masked_message = " ".join(words)

    log_entry = LogEntry(event_type=event_type, message=masked_message)
    db.session.add(log_entry)
    db.session.commit()


def send_deletion_email(email, delete_link):
    if not ADMIN_MAILGUN_API_KEY:
        logger.error("Admin Mailgun API key is not set.")
        return

    response = requests.post(
        f"https://api.mailgun.net/v3/{MAILGUN_DOMAIN}/messages",
        auth=("api", ADMIN_MAILGUN_API_KEY),
        data={
            "from": f"Quacker Admin <no-reply@{MAILGUN_DOMAIN}>",
            "to": [email],
            "subject": "Confirm Blog Disable",
            "text": f"To confirm disabling your blog's newsletters, click the link below:\n\n{delete_link}"
        }
    )

    if response.status_code != 200:
        logger.error(f"Failed to send deletion email to {email}: {response.text}")
    else:
        logger.info(f"Deletion email sent to {email}.")

def garbage_collect_invite_codes():
    now = datetime.datetime.now()
    expired_codes = InviteCode.query.filter(
        (InviteCode.expires_at < now) | (InviteCode.used == True)
    ).all()

    for code in expired_codes:
        db.session.delete(code)
    db.session.commit()

    logger.info(f"Garbage collected {len(expired_codes)} expired invite codes.")

def send_email(blog, recipient, title, link, unsubscribe_link):
    """Send a newsletter email to a subscriber."""
    account = Account.query.get(blog.account_id)
    if not account:
        logger.error("No associated account found for the blog.")
        return

    # Prepare the shame message if donations are not made
    shame_message = ""
    if SHOW_SHAME_MESSAGE and not blog.donated_date:
        shame_message = """
        <p style="color: red; font-size: 12px; text-align: center; margin-top: 20px;">
            This newsletter is powered by <a href="https://quacker.eu" style="color: red; text-decoration: none;">Quacker</a>,
            but the blog owner has not donated. For shame!
        </p>
        """

    # Generate the email body
    email_body = generate_email_body(
        title=title,
        description=f"Check out the latest update from {blog.home_url}.",
        article_link=link,
        blog_title=blog.home_url,
        unsubscribe_link=unsubscribe_link,
    ) + shame_message

    # Mask the recipient email for logging
    masked_email = mask_email(recipient)

    try:
        # Send the email via Mailgun
        response = requests.post(
            f"https://api.mailgun.net/v3/{account.mailgun_domain}/messages",
            auth=("api", account.mailgun_api_key),
            data={
                "from": f"{blog.home_url} <{account.newsletter_email}>",
                "to": [recipient],
                "subject": title,
                "html": email_body,
            },
        )

        # Log the result of the email send
        if response.status_code != 200:
            logger.error(f"Failed to send email to {masked_email}: {response.text}")
        else:
            log_event(
                "newsletter_sent",
                f"Newsletter '{title[:30]}...' sent for {blog.home_url} to subscriber {masked_email}."
            )
            logger.info(f"Email sent to {masked_email}.")
    except Exception as e:
        logger.error(f"Error while sending email to {masked_email}: {e}")


def check_rss_feeds():
    with app.app_context():
        today = datetime.date.today()
        two_days_ago = today - datetime.timedelta(days=2)

        blogs = Blog.query.all()
        for blog in blogs:
            try:
                response = requests.get(blog.rss_feed, timeout=5)
                if response.status_code != 200:
                    logger.warning(f"RSS feed fetch failed for blog {blog.id}: {blog.rss_feed}")
                    continue

                soup = BeautifulSoup(response.text, 'xml')
                items = soup.find_all('item')

                seen_articles = set()

                for item in items:
                    title = item.title.text
                    link = item.link.text
                    pub_date = item.pubDate.text if item.pubDate else None

                    if pub_date:
                        pub_date_utc = parser.parse(pub_date).astimezone(utc)
                    else:
                        logger.warning(f"Missing publication date for article: {link}")
                        continue

                    # Ignore articles older than 2 days
                    if pub_date_utc.date() < two_days_ago:
                        continue

                    # Ignore duplicate articles in the same feed
                    if link in seen_articles:
                        logger.info(f"Duplicate article in RSS feed ignored: {link}")
                        continue
                    seen_articles.add(link)

                    # Skip if the article was already sent
                    already_sent = db.session.query(SentArticle).filter_by(
                        blog_id=blog.id,
                        article_link=link
                    ).first()
                    if already_sent:
                        logger.info(f"Article already sent: {link}")
                        continue

                    # Send the email to subscribers
                    subscribers = Subscriber.query.filter_by(blog_id=blog.id).all()
                    for subscriber in subscribers:
                        unsubscribe_link = f"{APP_HOST_URL}unsubscribe?email={subscriber.email}&blog_id={blog.id}"
                        clean_title = title.replace("http://", "").replace("https://", "")
                        send_email(blog, subscriber.email, clean_title, link, unsubscribe_link)

                    # Add the article to the SentArticle table
                    try:
                        sent_article = SentArticle(
                            blog_id=blog.id,
                            article_link=link,
                            date_sent=today
                        )
                        db.session.add(sent_article)
                        db.session.commit()
                        logger.info(f"Article marked as sent: {link}")
                    except IntegrityError as e:
                        db.session.rollback()
                        logger.warning(f"IntegrityError when adding sent article {link}: {e}")
                        continue

            except Exception as e:
                db.session.rollback()
                logger.error(f"Error polling RSS feed for blog {blog.id}: {e}")


def generate_email_body(title, description, article_link, blog_title, unsubscribe_link, additional_note=""):
    return f"""
    <div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
        <h1 style="text-align: center; color: #2c3e50;">{title}</h1>
        <p style="text-align: center; margin: 20px 0;">{description}</p>
        <p style="text-align: center;">
            <a href="{article_link}" style="background-color: #3498db; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">
                {blog_title}
            </a>
        </p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        {additional_note}
        <p style="font-size: 12px; text-align: center; color: #777;">
            <a href="{unsubscribe_link}" style="color: #777;">Unsubscribe</a>
        </p>
    </div>
    """

def mask_email(email):
    """Mask email addresses to show only the first and last characters of the local part."""
    if "@" not in email:
        return email  # Return as is if it's not a valid email
    local, domain = email.split("@", 1)
    if len(local) <= 2:
        masked_local = local[0] + "*"
    else:
        masked_local = local[0] + "*" * (len(local) - 2) + local[-1]
    return f"{masked_local}****"


def garbage_collect_sent_articles():
    one_week_ago = datetime.date.today() - datetime.timedelta(days=7)
    old_articles = SentArticle.query.filter(SentArticle.date_sent < one_week_ago).delete()
    db.session.commit()
    logger.info(f"Garbage collection complete: Removed {old_articles} old sent articles.")

def garbage_collect_logs():
    threshold = datetime.datetime.utcnow() - datetime.timedelta(days=30)
    LogEntry.query.filter(LogEntry.timestamp < threshold).delete()
    db.session.commit()
    logger.info("Old logs removed.")


def start_scheduler():
    scheduler.add_job(garbage_collect_logs, "cron", hour=1)
    scheduler.add_job(garbage_collect_invite_codes, "cron", hour=0)
    scheduler.add_job(garbage_collect_sent_articles, "cron", hour=2)
    scheduler.add_job(check_rss_feeds, "interval", minutes=15)
    scheduler.start()


setup_database()
start_scheduler() 

if __name__ == '__main__':
    app.run(debug=True)
    
