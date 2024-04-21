import requests
import os

def send_html_files(directory):
    url = 'http://localhost:9090/api/holidays/import'
    # File to keep track of successfully uploaded files
    log_file_path = os.path.join(directory, 'upload_log.txt')

    # Try to load already uploaded file names into a set
    try:
        with open(log_file_path, 'r') as log_file:
            uploaded_files = set(log_file.read().splitlines())
    except FileNotFoundError:
        uploaded_files = set()

    for filename in os.listdir(directory):
        if filename.endswith('.html') and filename not in uploaded_files:
            filepath = os.path.join(directory, filename)
            with open(filepath, 'rb') as file:
                files = {'file': (filename, file)}
                response = requests.post(url, files=files)
                if response.status_code == 200:
                    print(f"Successfully uploaded {filename}")
                    # Add filename to uploaded files and update log file
                    uploaded_files.add(filename)
                    with open(log_file_path, 'a') as log_file:
                        log_file.write(filename + '\n')
                else:
                    print(f"Failed to upload {filename}: {response.content}")

if __name__ == "__main__":
    send_html_files("/app/holidays_2024")
