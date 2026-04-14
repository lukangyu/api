export function getToken(): string {
  return localStorage.getItem('token') || ''
}

export function setToken(token: string) {
  localStorage.setItem('token', token)
}

export function clearToken() {
  localStorage.removeItem('token')
}

export function isLoggedIn(): boolean {
  return !!getToken()
}
