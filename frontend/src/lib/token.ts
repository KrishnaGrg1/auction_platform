import Cookies from 'js-cookie'
import { getRequest } from '@tanstack/react-start/server'

const TOKEN_KEY = 'token'

/**
 * Set authentication token in cookies (client-side only)
 */
export const setAuth = (token: string) => {
  if (typeof window !== 'undefined') {
    Cookies.set(TOKEN_KEY, token, {
      expires: 7, // 7 days
      sameSite: 'strict',
      secure: import.meta.env.PROD, // Only secure in production
      path: '/',
    })
  }
}

/**
 * Get authentication token from cookies (client-side only)
 */
export function getAuthToken(): string | null {
  const request = getRequest()

  const cookie = request.headers.get('cookie')

  return (
    cookie
      ?.split('; ')
      .find((c) => c.startsWith('token='))
      ?.slice('token='.length) ?? null
  )
}

/**
 * Remove authentication token and redirect to login
 */
export const logout = () => {
  if (typeof window !== 'undefined') {
    Cookies.remove(TOKEN_KEY, { path: '/' })
    window.location.href = '/login'
  }
}

/**
 * Check if user is authenticated
 */
export const isAuthenticated = () => {
  const token = getAuthToken()
  return token !== null && token !== ''
}
