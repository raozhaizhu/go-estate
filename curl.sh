curl -X POST "http://localhost:9000/avatars/" \
  -F "Content-Type=image/png" \
  -F "bucket=avatars" \
  -F "key=18e90ca2-d225-4d16-9502-ffa5867a9af9.png" \
  -F "policy=eyJleHBpcmF0aW9uIjoiMjAyNi0wNy0yNlQwOTo0MjoxOS41OTdaIiwiY29uZGl0aW9ucyI6W1siZXEiLCIkYnVja2V0IiwiYXZhdGFycyJdLFsiZXEiLCIka2V5IiwiMThlOTBjYTItZDIyNS00ZDE2LTk1MDItZmZhNTg2N2E5YWY5LnBuZyJdLFsiZXEiLCIkQ29udGVudC1UeXBlIiwiaW1hZ2UvcG5nIl0sWyJlcSIsIiR4LWFtei1kYXRlIiwiMjAyNjA3MjZUMDkzNzE5WiJdLFsiZXEiLCIkeC1hbXotYWxnb3JpdGhtIiwiQVdTNC1ITUFDLVNIQTI1NiJdLFsiZXEiLCIkeC1hbXotY3JlZGVudGlhbCIsImFkbWluLzIwMjYwNzI2L3VzLWVhc3QtMS9zMy9hd3M0X3JlcXVlc3QiXSxbImNvbnRlbnQtbGVuZ3RoLXJhbmdlIiwgMTAyNCwgNTI0Mjg4MF1dfQ==" \
  -F "x-amz-algorithm=AWS4-HMAC-SHA256" \
  -F "x-amz-credential=admin/20260726/us-east-1/s3/aws4_request" \
  -F "x-amz-date=20260726T093719Z" \
  -F "x-amz-signature=f7df5ee76abb2c119f9930acdb077a7a7c5c0cacf846e733ead09567e1969b1c" \
  -F "file=@test.png"