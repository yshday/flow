/**
 * useFlowAuth Hook
 * 호스트 앱의 사용자 정보를 Flow 토큰으로 교환하는 인증 브릿지 훅
 */

import { useState, useCallback, useEffect } from 'react'
import axios from 'axios'
import type { FlowUser } from '../providers/FlowAuthProvider'

interface TokenExchangeRequest {
  provider: string
  external_id: string
  email: string
  username: string
  name?: string
  avatar_url?: string
}

interface TokenExchangeResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: FlowUser & {
    external_id?: string
    external_provider?: string
  }
  created: boolean
}

interface UseFlowAuthOptions {
  /** Flow API 서버 URL */
  apiBaseUrl: string
  /** SSO Provider 이름 (예: 'jmember') */
  provider: string
  /** 자동으로 토큰 교환 실행 여부 */
  autoExchange?: boolean
}

interface UseFlowAuthReturn {
  /** Flow 액세스 토큰 */
  accessToken: string | null
  /** Flow 사용자 정보 */
  user: FlowUser | null
  /** 로딩 상태 */
  isLoading: boolean
  /** 에러 */
  error: Error | null
  /** 토큰 교환 함수 */
  exchangeToken: (externalUser: {
    id: string
    email: string
    username: string
    name?: string
    avatar_url?: string
  }) => Promise<TokenExchangeResponse>
  /** 인증 초기화 */
  reset: () => void
}

/**
 * Flow 인증 브릿지 훅
 * 호스트 앱의 사용자 정보를 Flow 토큰으로 교환합니다.
 *
 * @example
 * ```tsx
 * const { accessToken, user, exchangeToken } = useFlowAuth({
 *   apiBaseUrl: 'http://localhost:8080/api/v1',
 *   provider: 'jmember',
 * })
 *
 * // 호스트 앱의 사용자 정보로 토큰 교환
 * useEffect(() => {
 *   if (hostUser && !accessToken) {
 *     exchangeToken({
 *       id: hostUser.id,
 *       email: hostUser.email,
 *       username: hostUser.username,
 *       name: hostUser.name,
 *     })
 *   }
 * }, [hostUser])
 *
 * // FlowIssueTracker에 전달
 * <FlowIssueTracker
 *   config={{ apiBaseUrl: 'http://localhost:8080/api/v1' }}
 *   user={user}
 *   accessToken={accessToken}
 * />
 * ```
 */
export function useFlowAuth(options: UseFlowAuthOptions): UseFlowAuthReturn {
  const { apiBaseUrl, provider } = options

  const [accessToken, setAccessToken] = useState<string | null>(null)
  const [user, setUser] = useState<FlowUser | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const exchangeToken = useCallback(
    async (externalUser: {
      id: string
      email: string
      username: string
      name?: string
      avatar_url?: string
    }): Promise<TokenExchangeResponse> => {
      setIsLoading(true)
      setError(null)

      try {
        const request: TokenExchangeRequest = {
          provider,
          external_id: externalUser.id,
          email: externalUser.email,
          username: externalUser.username,
          name: externalUser.name,
          avatar_url: externalUser.avatar_url,
        }

        const response = await axios.post<TokenExchangeResponse>(
          `${apiBaseUrl}/auth/token-exchange`,
          request,
          {
            headers: {
              'Content-Type': 'application/json',
            },
          }
        )

        const { access_token, user: flowUser } = response.data

        setAccessToken(access_token)
        setUser(flowUser)

        // 로컬 스토리지에 토큰 저장 (선택적)
        if (typeof window !== 'undefined') {
          localStorage.setItem('flow_access_token', access_token)
          localStorage.setItem('flow_refresh_token', response.data.refresh_token)
        }

        return response.data
      } catch (err) {
        const error = err instanceof Error ? err : new Error('Token exchange failed')
        setError(error)
        throw error
      } finally {
        setIsLoading(false)
      }
    },
    [apiBaseUrl, provider]
  )

  const reset = useCallback(() => {
    setAccessToken(null)
    setUser(null)
    setError(null)

    if (typeof window !== 'undefined') {
      localStorage.removeItem('flow_access_token')
      localStorage.removeItem('flow_refresh_token')
    }
  }, [])

  // 초기화 시 로컬 스토리지에서 토큰 복원
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const storedToken = localStorage.getItem('flow_access_token')
      if (storedToken) {
        setAccessToken(storedToken)
        // TODO: 토큰 유효성 검증 및 사용자 정보 가져오기
      }
    }
  }, [])

  return {
    accessToken,
    user,
    isLoading,
    error,
    exchangeToken,
    reset,
  }
}
