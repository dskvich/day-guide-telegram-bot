# Import Holidays

## Overview
This document outlines the process for importing holiday data from the "What Holiday Is It Today" website into our system.
The data is scraped, transferred, and processed to keep our holiday records up-to-date.

## Steps

### 1. Data Scraping
Use the `https://github.com/dskvich/holyscrape` tool to scrape holiday data from the "What Holiday Is It Today" website.
This will generate HTML files for each day of the year, resulting in either 356 or 366 files depending on the year.

### 2. Transfer Files
Copy the scraped HTML files to the `/app/holiday_2024` and the script `script_import_holidays.py` to the `/app` directory to the server using WinSCP.

### 3. Run Import Script
Run the script_import_holidays.py. Python 3.10.12 is currently used on the server.

```bash
python3 script_import_holidays.py
```

The script sends HTML files from the `/app/holiday_2024` directory to the server at the endpoint `/api/holidays/import`
and logs the names of successfully uploaded files.  It skips files that have already been uploaded, ensuring each file is only sent once.
If a file uploads successfully, its name is recorded in a log file. If an upload fails, it reports the failure.

### 4. Verify Uploads
To verify the number of files successfully uploaded, run the following command:

```bash
wc -l holidays_2024/upload_log.txt
```

This command counts the number of lines in the log file, which corresponds to the number of successfully uploaded files.