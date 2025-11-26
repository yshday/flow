/**
 * Flow API Client
 * 호스트 앱에서 주입받은 토큰을 사용하는 API 클라이언트
 */

import axios from 'axios'
import type { AxiosInstance } from 'axios'

let flowApiClient: AxiosInstance | null = null
let currentToken: string | null = null
let currentBaseUrl: string | null = null

/**
 * Flow API 클라이언트 초기화
 */
export function initFlowClient(baseUrl: string, accessToken: string | null): AxiosInstance {
  // 동일한 설정이면 기존 인스턴스 재사용
  if (flowApiClient && currentBaseUrl === baseUrl && currentToken === accessToken) {
    return flowApiClient
  }

  currentBaseUrl = baseUrl
  currentToken = accessToken

  flowApiClient = axios.create({
    baseURL: baseUrl,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // Request interceptor - Add JWT token
  flowApiClient.interceptors.request.use(
    (config) => {
      if (accessToken) {
        config.headers.Authorization = `Bearer ${accessToken}`
      }
      return config
    },
    (error) => Promise.reject(error)
  )

  return flowApiClient
}

/**
 * 현재 Flow API 클라이언트 가져오기
 */
export function getFlowClient(): AxiosInstance {
  if (!flowApiClient) {
    throw new Error('Flow client not initialized. Call initFlowClient first.')
  }
  return flowApiClient
}

/**
 * 토큰 업데이트
 */
export function updateFlowToken(accessToken: string | null) {
  currentToken = accessToken
  if (flowApiClient && currentBaseUrl) {
    initFlowClient(currentBaseUrl, accessToken)
  }
}
