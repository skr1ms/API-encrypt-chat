$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDgzNjc3MjIsImlhdCI6MTc0ODI4MTMyMiwianRpIjoiMDc0OWY5MDUtNGE2YS00MjUwLWI3MWUtM2YzN2UyNjM4YmVjIiwidXNlcl9pZCI6NH0._fboEuEKGfglU0Ne1o83F-cJhZ3tPHt3hGEnr1lGLGY"

$headers = @{
    'Authorization' = "Bearer $token"
    'Content-Type' = 'application/json'
}

try {
    $response = Invoke-RestMethod -Uri 'http://localhost:8080/api/v1/users/search?q=test&limit=10' -Method Get -Headers $headers
    Write-Output "Success!"
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Output "Error: $($_.Exception.Message)"
    Write-Output "Status Code: $($_.Exception.Response.StatusCode.value__)"
}
