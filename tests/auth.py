import requests
import jwt

BASE_URL = "http://localhost:8000/"

def test_signup_endpoint():
    data = {
        "username": "testuser",
        "password": "testpassword"
    }

    response = requests.post(f"{BASE_URL}/signup",json=data)

    assert response.status_code == 201, f"Expected 201, got {response.status_code}"

    cookies = response.cookies
    assert "ACCESS_TOKEN" in cookies, "Access token missing on signing up"
    assert "REFRESH_TOKEN" in cookies, "Refresh token missing on signing up"

    access_token = cookies.get(name="ACCESS_TOKEN")
    refresh_token = cookies.get(name="REFRESH_TOKEN")
    assert access_token is not None and len(access_token) > 0, "Access token is empty"
    assert refresh_token is not None and len(refresh_token) > 0, "Refresh token is empty"

    assert decode_jwt(access_token) is not None, "Access token not in correct format"
    assert decode_jwt(refresh_token) is not None, "Refresh token not in correct format"

def decode_jwt(token):
    try:
        payload = jwt.decode(token, options={"verify_signature": False})
        return payload
    except jwt.DecodeError:
        return None
