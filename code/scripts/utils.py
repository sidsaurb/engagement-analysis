import threading
import smtplib
from email.mime.text import MIMEText
from twilio.rest import TwilioRestClient
from twilio import TwilioException

GMAIL_SMTP = 'smtp.gmail.com'
GMAIL_SMTP_PORT = 587
GMAIL_ACCOUNT = 'gpahal.website@gmail.com'
GMAIL_ACCOUNT_PASSWORD = 'homeiitkacingpahal'

ACCOUNT_SID = "AC50e635175981074ce8faab35e1aa79e3"
AUTH_TOKEN = "36eb85f02a916730b2bb7686aa9255b0"

TWILIO_CLIENT = TwilioRestClient(ACCOUNT_SID, AUTH_TOKEN)


class EmailThread(threading.Thread):
    def __init__(self, to_email, subject, message):
        super(EmailThread, self).__init__()
        self.email_message = MIMEText(message)
        self.email_message['From'] = GMAIL_ACCOUNT
        self.email_message['To'] = to_email
        self.email_message['Subject'] = subject

    def run(self):
        try:
            smtp_conn = smtplib.SMTP(GMAIL_SMTP, GMAIL_SMTP_PORT)

            smtp_conn.ehlo()
            smtp_conn.starttls()
            smtp_conn.ehlo()

            smtp_conn.login(user=GMAIL_ACCOUNT, password=GMAIL_ACCOUNT_PASSWORD)

            smtp_conn.send_message(self.email_message)

            smtp_conn.quit()
            return True
        except:
            return False


class SMSThread(threading.Thread):
    def __init__(self, phone_no, message):
        super(SMSThread, self).__init__()
        self.phone_no = phone_no
        self.message = message

    def run(self):
        try:
            TWILIO_CLIENT.messages.create(
                    to=self.phone_no,
                    from_="+12016453730",
                    body=self.message,
            )

            return True
        except:
            return False
