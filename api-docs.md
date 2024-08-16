
# API Documentation

## Overview

The `storemypdf` API provides endpoints for uploading, retrieving, and downloading files. This document describes how to use the API.

**Base URL:**  
`https://storemypdf.com/api`

## Authentication

API requests must include an API key for authentication. You can provide the API key as a query parameter.

- **Query Parameter Authentication:** 
  - Append `?key=YOUR_API_KEY` to the endpoint URL.

Replace `YOUR_API_KEY` with your actual API key.

## Endpoints

### 1. Upload File

- **Endpoint:** `/files/upload`
- **Method:** `POST`
- **Description:** Uploads a new file to the server.
- **Headers:**
  - `Content-Type: multipart/form-data`
- **Parameters:**
  - **Body (Form Data):**
    - `file` (file, required): The file to upload.
    - `delete_after` (integer, optional): Time in hours after which the file will be automatically deleted. Valid values are 0, 1, 2, 3, 4. Default is 0 (do not delete).
    - `description` (string, optional): A description for the file.
- **Response:**
  - **Success (`200 OK`):**
    ```json
    {
      "success": true,
      "file_id": "unique-file-id",
      "message": "File uploaded successfully."
    }
    ```
  - **Error (`400 Bad Request`):**
    ```json
    {
      "success": false,
      "message": "Invalid file format or missing file."
    }
    ```

### 2. Get File

- **Endpoint:** `/files/{file_id}`
- **Method:** `GET`
- **Description:** Retrieves information about a file by its unique ID.
- **Headers:**
  - `Accept: application/json`
- **Parameters:**
  - **Path:**
    - `file_id` (string, required): The unique identifier of the file.
- **Response:**
  - **Success (`200 OK`):**
    ```json
    {
      "success": true,
      "file": {
        "id": "unique-file-id",
        "name": "filename.pdf",
        "url": "https://storemypdf/api/files/unique-file-id/download",
        "description": "Optional description of the file",
        "delete_after": 0
      }
    }
    ```
  - **Error (`404 Not Found`):**
    ```json
    {
      "success": false,
      "message": "File not found."
    }
    ```

### 3. Download File

- **Endpoint:** `/files/{file_id}/download`
- **Method:** `GET`
- **Description:** Downloads a file by its unique ID.
- **Headers:**
  - `Accept: application/octet-stream`
- **Parameters:**
  - **Path:**
    - `file_id` (string, required): The unique identifier of the file.
- **Response:**
  - **Success (`200 OK`):** The file content is returned in the response body.
  - **Error (`404 Not Found`):**
    ```json
    {
      "success": false,
      "message": "File not found."
    }
    ```

## Error Codes

- **400 Bad Request:** The request was invalid. This could be due to missing parameters or invalid values.
- **404 Not Found:** The requested resource could not be found. This may occur if the file ID does not exist.
- **500 Internal Server Error:** An unexpected error occurred on the server.

## Rate Limiting

- **Maximum Requests:** 1000 requests per hour per IP address.
- **Rate Limit Exceeded:** `429 Too Many Requests`

## Examples

### Upload File

**Request:**
```bash
curl -X POST "https://storemypdf/api/files/upload" \
-H "Content-Type: multipart/form-data" \
-F "file=@/path/to/file.pdf" \
-F "delete_after=1" \
-F "description=Sample PDF file"
