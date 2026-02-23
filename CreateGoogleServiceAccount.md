## Enable the APIs 
* Before creating credentials, you must enable the necessary services for your project: 
* Go to the Google Cloud Console.
* Create a new project or select an existing one.
* Go to APIs & Services > Library.
* Search for and Enable both the Google Drive API (needed for Docs) and the Google Sheets API. 
## Generate the credentials.json
### Service Account (Recommended for simple scripts) 
* Go to IAM & Admin > Service Accounts.
* Click Create Service Account, give it a name, and click Done.
* Click on the email address of your new service account.
* Go to the Keys tab, click Add Key > Create new key.
* Select JSON and click Create. A file will download automatically—this is your credentials.json.
* Crucial: To access specific files, you must open the Doc or Sheet and "Share" it with the service account's email address. 
