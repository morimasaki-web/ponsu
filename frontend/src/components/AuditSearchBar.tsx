import React, { useState, useEffect } from "react"
import { useQuery } from 'urql'
import { useSearchParams } from 'react-router-dom'
import { AuditLogsQuery, AuditLogsDocument } from "../gql/graphql"
import { useDebounce } from "../hooks/useDebounce"
import { appMode } from '../graphql'
import { DemoAuditLog } from '../demoData'

/**
 * 日付文字列（YYYY-MM-DD）をRFC3339形式に変換
 * GraphQLのTime型はRFC3339Nano形式を期待するため
 */
function toRFC3339(dateString: string): string {
  if (!dateString) return ""
  // YYYY-MM-DD → YYYY-MM-DDT00:00:00Z
  return `${dateString}T00:00:00Z`
}

interface AuditSearchBarProps {
  onSearchResultsChange?: (results: AuditLogsQuery['auditLogs'] | null) => void
  demoData?: DemoAuditLog[]
}

export function AuditSearchBar({ 
  onSearchResultsChange,
  demoData = []
}: AuditSearchBarProps) {
  const [searchParams, setSearchParams] = useSearchParams()

  // URLパラメータから初期値を取得
  const requestID = searchParams.get('requestID') // 申請IDでのフィルタ（RequestDetailからのリンク用）
  const [actorUserID, setActorUserID] = useState(searchParams.get('actorUserID') || "")
  const [action, setAction] = useState(searchParams.get('action') || "")
  const [occurredAtStart, setOccurredAtStart] = useState(searchParams.get('start') || "")
  const [occurredAtEnd, setOccurredAtEnd] = useState(searchParams.get('end') || "")

  // デバウンス処理（入力後300ms待ってから検索）
  const debouncedActorUserID = useDebounce(actorUserID, 300)
  const debouncedAction = useDebounce(action, 300)
  const debouncedOccurredAtStart = useDebounce(occurredAtStart, 300)
  const debouncedOccurredAtEnd = useDebounce(occurredAtEnd, 300)

  // URLパラメータを更新（requestIDは除く - RequestDetailからのリンクで設定される）
  useEffect(() => {
    const params = new URLSearchParams()
    if (requestID) params.set('requestID', requestID) // requestIDは保持
    if (debouncedActorUserID) params.set('actorUserID', debouncedActorUserID)
    if (debouncedAction) params.set('action', debouncedAction)
    if (debouncedOccurredAtStart) params.set('start', debouncedOccurredAtStart)
    if (debouncedOccurredAtEnd) params.set('end', debouncedOccurredAtEnd)
    setSearchParams(params, { replace: true })
  }, [requestID, debouncedActorUserID, debouncedAction, debouncedOccurredAtStart, debouncedOccurredAtEnd, setSearchParams])

  // GraphQLクエリ（デバウンスされた値で自動的に検索）
  // デモモードの場合はクエリを実行しない
  // actorUserIDが8文字未満の場合もクエリを実行しない（UUID形式チェック）
  const shouldPauseQuery = 
    appMode === 'demo' || 
    (!!debouncedActorUserID && debouncedActorUserID.length < 8)
  
  const [{ data, fetching, error }] = useQuery({
    query: AuditLogsDocument,
    variables: {
      requestID: requestID || null,
      actorUserID: debouncedActorUserID || null,
      action: debouncedAction || null,
      occurredAtStart: debouncedOccurredAtStart ? toRFC3339(debouncedOccurredAtStart) : null,
      occurredAtEnd: debouncedOccurredAtEnd ? toRFC3339(debouncedOccurredAtEnd) : null,
      limit: 50,
      offset: 0,
    },
    requestPolicy: 'cache-and-network',
    pause: shouldPauseQuery,
  })

  // プロダクションモード: GraphQL検索結果を親コンポーネントに通知
  useEffect(() => {
    if (appMode === 'prod' && data?.auditLogs && onSearchResultsChange) {
      onSearchResultsChange(data.auditLogs)
    }
  }, [data?.auditLogs, onSearchResultsChange])

  // デモモード: クライアントサイドフィルタリング
  useEffect(() => {
    if (appMode === 'demo' && onSearchResultsChange) {
      const filtered = demoData.filter((log) => {
        // 申請IDでフィルタ
        if (requestID && log.requestID !== requestID) {
          return false
        }
        
        // 実行者IDでフィルタ
        if (debouncedActorUserID && !log.actorUserID?.includes(debouncedActorUserID)) {
          return false
        }
        
        // アクションでフィルタ
        if (debouncedAction && log.action !== debouncedAction) {
          return false
        }
        
        // 発生日時範囲でフィルタ
        if (debouncedOccurredAtStart && log.occurredAt < toRFC3339(debouncedOccurredAtStart)) {
          return false
        }
        if (debouncedOccurredAtEnd) {
          const endDate = new Date(debouncedOccurredAtEnd)
          endDate.setDate(endDate.getDate() + 1) // 終了日を含めるため+1日
          if (log.occurredAt >= endDate.toISOString()) {
            return false
          }
        }
        
        return true
      })
      
      // フィルタ条件がある場合のみ検索結果として通知
      if (requestID || debouncedActorUserID || debouncedAction || debouncedOccurredAtStart || debouncedOccurredAtEnd) {
        // DemoAuditLog を GraphQL型に変換
        const convertedResults = filtered.map(log => ({
          __typename: 'RequestAudit' as const,
          id: log.id,
          actorUserID: log.actorUserID,
          action: log.action,
          data: log.data,
          occurredAt: log.occurredAt
        }))
        onSearchResultsChange(convertedResults)
      } else {
        onSearchResultsChange(null) // フィルタなし = 全件表示
      }
    }
  }, [requestID, debouncedActorUserID, debouncedAction, debouncedOccurredAtStart, debouncedOccurredAtEnd, demoData, onSearchResultsChange])

  // リセット機能
  const handleReset = () => {
    setActorUserID("")
    setAction("")
    setOccurredAtStart("")
    setOccurredAtEnd("")
    // 親コンポーネントに全件表示に戻すよう通知
    if (onSearchResultsChange) {
      onSearchResultsChange(null)
    }
  }

  return (
    <div className="requestSearchBarContainer">
      <div className="requestSearchBar">
        <div className="searchInputGroup">
          <svg className="searchIcon" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="8"></circle>
            <path d="m21 21-4.35-4.35"></path>
          </svg>
          <input 
            type="text" 
            name="actorUserID" 
            placeholder="実行者ID（8文字以上）で検索..." 
            value={actorUserID}
            onChange={(e) => setActorUserID(e.target.value)}
            className="searchInput searchInput--main"
          />
        </div>
        
        <select 
          name="action"
          value={action}
          onChange={(e) => setAction(e.target.value)}
          className="searchInput searchInput--select"
        >
          <option value="">すべてのアクション</option>
          <option value="submit">提出</option>
          <option value="approve">承認</option>
          <option value="reject">却下</option>
          <option value="return">差し戻し</option>
          <option value="resubmit">再提出</option>
        </select>
        
        <div className="searchInputGroup searchInputGroup--date">
          <input 
            type="date"
            name="occurredAtStart"
            value={occurredAtStart}
            onChange={(e) => setOccurredAtStart(e.target.value)}
            className="searchInput searchInput--date"
            title="発生日時開始"
          />
          <span className="dateSeparator">〜</span>
          <input 
            type="date"
            name="occurredAtEnd"
            value={occurredAtEnd}
            onChange={(e) => setOccurredAtEnd(e.target.value)}
            className="searchInput searchInput--date"
            title="発生日時終了"
          />
        </div>
        
        <button 
          type="button" 
          onClick={handleReset} 
          className="resetButton"
          title="検索条件をクリア"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M3 6h18"></path>
            <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
            <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
            <line x1="10" x2="10" y1="11" y2="17"></line>
            <line x1="14" x2="14" y1="11" y2="17"></line>
          </svg>
          クリア
        </button>
      </div>
      
      <div className="searchStatusBar">
        {appMode === 'prod' && fetching && (
          <div className="searchStatus">
            <svg className="spinner" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="12" x2="12" y1="2" y2="6"></line>
              <line x1="12" x2="12" y1="18" y2="22"></line>
              <line x1="4.93" x2="7.76" y1="4.93" y2="7.76"></line>
              <line x1="16.24" x2="19.07" y1="16.24" y2="19.07"></line>
              <line x1="2" x2="6" y1="12" y2="12"></line>
              <line x1="18" x2="22" y1="12" y2="12"></line>
              <line x1="4.93" x2="7.76" y1="19.07" y2="16.24"></line>
              <line x1="16.24" x2="19.07" y1="7.76" y2="4.93"></line>
            </svg>
            検索中...
          </div>
        )}
        
        {appMode === 'prod' && error && <div className="searchError">エラー: {error.message}</div>}
        
        {appMode === 'prod' && data?.auditLogs && (
          <div className="searchResultCount">
            {data.auditLogs.length}件の監査ログが見つかりました
          </div>
        )}
        
        {appMode === 'demo' && (debouncedActorUserID || debouncedAction || debouncedOccurredAtStart || debouncedOccurredAtEnd) && (
          <div className="searchResultCount">
            デモモードで検索中...
          </div>
        )}
      </div>
    </div>
  )
}