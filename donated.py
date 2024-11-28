import os
from sqlalchemy import create_engine, Column, Integer, String, DateTime, ForeignKey
from sqlalchemy.orm import sessionmaker, declarative_base
from datetime import datetime
from pytz import utc

# Load environment variables
DATABASE_URL = f"sqlite:///{os.path.join(os.path.abspath(os.path.dirname(__file__)), 'quacker.db')}"
engine = create_engine(DATABASE_URL)
Base = declarative_base()
Session = sessionmaker(bind=engine)
session = Session()

# Define the Account model
class Account(Base):
    __tablename__ = 'account'
    id = Column(Integer, primary_key=True)
    newsletter_email = Column(String(120), nullable=False)
    _mailgun_api_key = Column(String(255), nullable=True)
    mailgun_domain = Column(String(255), nullable=False)
    delete_token = Column(String(255), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

# Define the Blog model
class Blog(Base):
    __tablename__ = 'blog'
    id = Column(Integer, primary_key=True)
    account_id = Column(Integer, ForeignKey('account.id'), nullable=False)
    rss_feed = Column(String(255), unique=True, nullable=False)
    home_url = Column(String(255), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    last_email_sent = Column(DateTime, nullable=True)
    donated_date = Column(DateTime, nullable=True)

def mark_blog_as_donated(blog_id):
    """Mark a blog as donated by updating the donated_date."""
    blog = session.get(Blog, blog_id)  # Use `Session.get()` for SQLAlchemy 2.0+
    if not blog:
        print(f"No blog found with ID {blog_id}.")
        return

    blog.donated_date = datetime.now(utc)  # Corrected timezone-aware datetime
    session.commit()
    print(f"Blog with ID {blog_id} marked as donated.")

if __name__ == "__main__":
    import sys
    if len(sys.argv) != 2:
        print("Usage: python donated.py <blog_id>")
    else:
        try:
            blog_id = int(sys.argv[1])
            mark_blog_as_donated(blog_id)
        except ValueError:
            print("Error: Blog ID must be an integer.")
