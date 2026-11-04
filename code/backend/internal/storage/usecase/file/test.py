import base64
import hashlib
import os
from pathlib import Path

import requests


def sha256_hex(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def sha256_base64(data: bytes) -> str:
    return base64.b64encode(hashlib.sha256(data).digest()).decode("utf-8")


def detect_mime_type(file_path: str) -> str:
    return {
        ".pdf": "application/pdf",
        ".jpg": "image/jpeg",
        ".jpeg": "image/jpeg",
        ".png": "image/png",
        ".txt": "text/plain",
        ".doc": "application/msword",
        ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    }.get(Path(file_path).suffix.lower(), "application/octet-stream")


def init_upload(file_path: str, library_id: str, folder_id: str, auth_token: str, base_url: str):
    with open(file_path, "rb") as file_handle:
        file_data = file_handle.read()

    checksum_hex = sha256_hex(file_data)
    file_size = len(file_data)
    file_name = os.path.basename(file_path)
    mime_type = detect_mime_type(file_path)

    url = f"{base_url}/v1/storage/files/init"
    headers = {
        "accept": "application/json",
        "Authorization": f"Bearer {auth_token}",
        "Content-Type": "application/json",
    }
    payload = {
        "checksum_declared": checksum_hex,
        "folder_id": folder_id,
        "library_id": library_id,
        "mime_type": mime_type,
        "name": file_name,
        "size": file_size,
    }

    print("Init upload...")
    response = requests.post(url, headers=headers, json=payload, timeout=30)
    print(f"HTTP status: {response.status_code} {response.reason}")
    if response.status_code != 201:
        print(response.text)
        return None

    return response.json()


def presign_part(file_id: str, upload_id: str, part_number: int, checksum_b64: str, size: int,
                 auth_token: str, base_url: str) -> str:
    url = f"{base_url}/v1/storage/files/{file_id}/parts"
    headers = {
        "accept": "application/json",
        "Authorization": f"Bearer {auth_token}",
        "Content-Type": "application/json",
    }
    payload = {
        "upload_id": upload_id,
        "part_number": part_number,
        "checksum_sha256": checksum_b64,
        "size": size,
    }

    response = requests.post(url, headers=headers, json=payload, timeout=30)
    if response.status_code != 200:
        raise RuntimeError(f"Presign part failed: {response.status_code} {response.text}")

    return response.json()["url"]


def upload_part(presigned_url: str, part_data: bytes, checksum_b64: str) -> str:
    headers = {
        "Content-Length": str(len(part_data)),
        "x-amz-checksum-sha256": checksum_b64,
        "x-amz-checksum-algorithm": "SHA256",
    }
    response = requests.put(presigned_url, data=part_data, headers=headers, timeout=60)
    if response.status_code not in (200, 201):
        raise RuntimeError(f"Upload part failed: {response.status_code} {response.text}")

    return response.headers.get("ETag", "")


def confirm_upload(file_id: str, upload_id: str, completed_parts: list, auth_token: str, base_url: str) -> dict:
    url = f"{base_url}/v1/storage/files/{file_id}/complete"
    headers = {
        "accept": "application/json",
        "Authorization": f"Bearer {auth_token}",
        "Content-Type": "application/json",
    }
    payload = {
        "upload_id": upload_id,
        "completed_parts": completed_parts,
        "vectorize": False,
    }

    response = requests.post(url, headers=headers, json=payload, timeout=30)
    if response.status_code != 200:
        raise RuntimeError(f"Confirm upload failed: {response.status_code} {response.text}")

    return response.json()


def upload_multipart(file_path: str, file_info: dict, upload_id: str, auth_token: str, base_url: str,
                     part_size: int) -> list:
    completed_parts = []
    part_number = 1

    with open(file_path, "rb") as file_handle:
        while True:
            chunk = file_handle.read(part_size)
            if not chunk:
                break

            checksum_b64 = sha256_base64(chunk)
            presigned_url = presign_part(
                file_id=file_info["id"],
                upload_id=upload_id,
                part_number=part_number,
                checksum_b64=checksum_b64,
                size=len(chunk),
                auth_token=auth_token,
                base_url=base_url,
            )

            etag = upload_part(presigned_url, chunk, checksum_b64)
            completed_parts.append({"etag": etag, "part_number": part_number})
            print(f"Uploaded part {part_number}, size={len(chunk)}")
            part_number += 1

    return completed_parts


def upload_single(file_path: str, presigned_url: str, mime_type: str):
    with open(file_path, "rb") as file_handle:
        file_data = file_handle.read()

    checksum_b64 = sha256_base64(file_data)
    headers = {
        "Content-Type": mime_type,
        "Content-Length": str(len(file_data)),
        "x-amz-checksum-sha256": checksum_b64,
        "x-amz-checksum-algorithm": "SHA256",
    }
    response = requests.put(presigned_url, data=file_data, headers=headers, timeout=60)
    if response.status_code not in (200, 201):
        raise RuntimeError(f"Upload failed: {response.status_code} {response.text}")


def main():
    base_url = os.getenv("BASE_URL", "http://localhost:8080")
    file_path = os.getenv("FILE_PATH", "Maxim_HW.pdf")
    library_id = os.getenv("LIBRARY_ID", "d44b8d0e-68df-4f17-b87e-98fb858cb43b")
    folder_id = os.getenv("FOLDER_ID", "6df223b1-7f78-404f-b261-54f65fe8f291")
    auth_token = os.getenv("AUTH_TOKEN", "")
    part_size = int(os.getenv("PART_SIZE", str(5 * 1024 * 1024)))

    if not library_id or not auth_token:
        raise RuntimeError("LIBRARY_ID and AUTH_TOKEN are required")

    api_response = init_upload(file_path, library_id, folder_id, auth_token, base_url)
    if not api_response:
        raise RuntimeError("Init upload failed")

    file_info = api_response["File"]
    presigned_url = api_response.get("PresignURL")
    upload_id = api_response.get("UploadID", "")

    if presigned_url:
        upload_single(file_path, presigned_url, file_info["mime_type"])
        confirm_upload(file_info["id"], "", [], auth_token, base_url)
        print("Single upload completed")
        return

    if not upload_id:
        raise RuntimeError("Missing UploadID for multipart upload")

    completed_parts = upload_multipart(file_path, file_info, upload_id, auth_token, base_url, part_size)
    confirm_upload(file_info["id"], upload_id, completed_parts, auth_token, base_url)
    print("Multipart upload completed")


if __name__ == "__main__":
    main()