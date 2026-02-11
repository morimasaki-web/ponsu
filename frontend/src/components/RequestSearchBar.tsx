import React, { useState, useEffect } from "react"
import { useQuery } from 'urql'
import { useSearchParams } from 'react-router-dom'
import { SearchRequestsDocument, SearchRequestsQuery } from "../gql/graphql"
import { useDebounce } from "../hooks/useDebounce"

/**
 * 日付文字列（YYYY-MM-DD）をRFC3339形式に変換
 * GraphQLのTime型はRFC3339Nano形式を期待するため
 */
function toRFC3339(dateString: string): string {
  if (!dateString) return ""
  // YYYY-MM-DD → YYYY-MM-DDT00:00:00Z
  return `${dateString}T00:00:00Z`
}

interface RequestSearchBarProps {
  onSearchResultsChange?: (results: SearchRequestsQuery['searchRequests'] | null) => void
}

export function RequestSearchBar({ onSearchResultsChange }: RequestSearchBarProps) {
  const [searchParams, setSearchParams] = useSearchParams()

  // URLパラメータから初期値を取得
  const [title, setTitle] = useState(searchParams.get('title') || "")
  const [status, setStatus] = useState(searchParams.get('status') || "")
  const [createdAtStart, setCreatedAtStart] = useState(searchParams.get('start') || "")
  const [createdAtEnd, setCreatedAtEnd] = useState(searchParams.get('end') || "")

  // デバウンス処理（入力後300ms待ってから検索）
  const debouncedTitle = useDebounce(title, 300)
  const debouncedStatus = useDebounce(status, 300)
  const debouncedCreatedAtStart = useDebounce(createdAtStart, 300)
  const debouncedCreatedAtEnd = useDebounce(createdAtEnd, 300)

  // URLパラメータを更新
  useEffect(() => {
    const params = new URLSearchParams()
    if (debouncedTitle) params.set('title', debouncedTitle)
    if (debouncedStatus) params.set('status', debouncedStatus)
    if (debouncedCreatedAtStart) params.set('start', debouncedCreatedAtStart)
    if (debouncedCreatedAtEnd) params.set('end', debouncedCreatedAtEnd)
    setSearchParams(params, { replace: true })
  }, [debouncedTitle, debouncedStatus, debouncedCreatedAtStart, debouncedCreatedAtEnd, setSearchParams])

  // GraphQLクエリ（デバウンスされた値で自動的に検索）
  const [{ data, fetching, error }] = useQuery({
    query: SearchRequestsDocument,
    variables: {
      title: debouncedTitle || null,
      status: debouncedStatus || null,
      createdAtStart: debouncedCreatedAtStart ? toRFC3339(debouncedCreatedAtStart) : null,
      createdAtEnd: debouncedCreatedAtEnd ? toRFC3339(debouncedCreatedAtEnd) : null,
      limit: 50,
      offset: 0,
    },
    requestPolicy: 'cache-and-network',
  })

  // 検索結果が変わったら親コンポーネントに通知
  useEffect(() => {
    if (data?.searchRequests && onSearchResultsChange) {
      onSearchResultsChange(data.searchRequests)
    }
  }, [data?.searchRequests, onSearchResultsChange])

  // リセット機能
  const handleReset = () => {
    setTitle("")
    setStatus("")
    setCreatedAtStart("")
    setCreatedAtEnd("")
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
            name="title" 
            placeholder="タイトルで検索..." 
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="searchInput searchInput--main"
          />
        </div>
        
        <select 
          name="status"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          className="searchInput searchInput--select"
        >
          <option value="">すべてのステータス</option>
          <option value="pending">承認待ち</option>
          <option value="approved">承認済み</option>
          <option value="rejected">却下</option>
          <option value="returned">差し戻し</option>
        </select>
        
        <div className="searchInputGroup searchInputGroup--date">
          <input 
            type="date"
            name="createdAtStart"
            value={createdAtStart}
            onChange={(e) => setCreatedAtStart(e.target.value)}
            className="searchInput searchInput--date"
            title="作成日開始"
          />
          <span className="dateSeparator">〜</span>
          <input 
            type="date"
            name="createdAtEnd"
            value={createdAtEnd}
            onChange={(e) => setCreatedAtEnd(e.target.value)}
            className="searchInput searchInput--date"
            title="作成日終了"
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
        {fetching && (
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
        
        {error && <div className="searchError">⚠️ エラー: {error.message}</div>}
        
        {data?.searchRequests && (
          <div className="searchResultCount">
            📋 {data.searchRequests.length}件の申請が見つかりました
          </div>
        )}
      </div>
    </div>
  )
}