# API Documentation
## Overview
The `storemypdf` API provides endpoints for uploading, retrieving, and downloading files. This document describes how to use the API.

**Base URL:**

`https://storemypdf.com/api/`

## Authentication
API requests must include an API key for authentication. You can provide the API key as a query parameter.
-  **Query Parameter Authentication:**
- Append `?key=YOUR_API_KEY` to the endpoint URL.

Replace `YOUR_API_KEY` with your actual API key. You can get your [API key here.](/u/api-keys)
## Endpoints
### 1. Upload File
-  **Endpoint:**  `/files/upload`

-  **Method:**  `POST`

-  **Description:** Uploads a new file to the server.

-  **Headers:**

-  `Content-Type: multipart/form-data`

-  **Parameters:**

-  **Body (Form Data):**

-  `file` (file, required): The file to upload.

-  `delete_after` (integer, optional): Time in hours after which the file will be automatically deleted. Valid values are 0, 1, 2, 3, 4. Default is 0 (do not delete).

-  `description` (string, optional): A description for the file.

-  **Response:**

-  **Success (`200 OK`):**

```json
{
"success": true,
"file_id": "unique-file-id",
"url": "http://storemypdf.com/f/unique-file-id",
"url_viewer": "https://files.storemypdf.com/unique-file-id",
"message": "File uploaded successfully"
}

```
-  **Error (`400 Bad Request`):**
```json
{
"success": false,
"message": "Invalid file format or missing file."
}
```
