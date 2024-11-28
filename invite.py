import os
import datetime
from uuid import uuid4
from dotenv import load_dotenv
from sqlalchemy import create_engine, Column, Integer, String, DateTime, Boolean
from sqlalchemy.orm import sessionmaker, declarative_base

# Load environment variables
load_dotenv()

# Define database URL and SQLAlchemy setup
DATABASE_URL = f"sqlite:///{os.path.join(os.path.abspath(os.path.dirname(__file__)), 'quacker.db')}"
engine = create_engine(DATABASE_URL)
Base = declarative_base()
Session = sessionmaker(bind=engine)
session = Session()

# Define the InviteCode model
class InviteCode(Base):
    __tablename__ = 'invite_code'
    id = Column(Integer, primary_key=True)
    code = Column(String(50), unique=True, nullable=False)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    used = Column(Boolean, default=False)
    expires_at = Column(DateTime, nullable=False)

# Generate invite code and insert it into the database
def generate_invite_code():
    code = str(uuid4())[:8]
    expiration_date = datetime.datetime.now() + datetime.timedelta(days=7)

    new_code = InviteCode(
        code=code,
        expires_at=expiration_date
    )
    session.add(new_code)
    session.commit()

    print(f"Invite Code: {code} (Expires: {expiration_date})")

if __name__ == "__main__":
    # Ensure the database schema is initialized
    Base.metadata.create_all(engine)
    generate_invite_code()
