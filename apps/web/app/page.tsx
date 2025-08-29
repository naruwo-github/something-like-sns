'use client'
import { useEffect, useState } from 'react'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'

type Post = {
  id: string | number
  authorUserId: string | number
  body: string
  createdAt: string
  likedByMe: boolean
  likeCount: number
  commentCount: number
}

async function call<TReq extends object, TRes>(path: string, req: TReq, headers: Record<string,string>): Promise<TRes> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      ...headers,
    },
    body: JSON.stringify(req ?? {}),
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new Error(`HTTP ${res.status} ${text}`)
  }
  return res.json() as Promise<TRes>
}

export default function Page() {
  const [tenantSlug] = useState('acme')
  const [user] = useState('u_alice')
  const [posts, setPosts] = useState<Post[]>([])
  const [newBody, setNewBody] = useState('')

  useEffect(() => {
    const headers: Record<string, string> = { 'X-Tenant': tenantSlug, 'X-User': user }
    ;(async () => {
      await call('/sns.v1.TenantService/GetMe', {}, headers)
      const res = await call<{},{ items: Post[] }>('/sns.v1.TimelineService/ListFeed', {}, headers)
      setPosts(res.items)
    })()
  }, [tenantSlug, user])

  const createPost = async () => {
    const headers: Record<string, string> = { 'X-Tenant': tenantSlug, 'X-User': user }
    const res = await call<{ body: string }, { post: Post }>('/sns.v1.TimelineService/CreatePost', { body: newBody }, headers)
    setPosts([res.post, ...posts])
    setNewBody('')
  }

  const toggleLike = async (postId: number) => {
    const headers: Record<string, string> = { 'X-Tenant': tenantSlug, 'X-User': user }
    const r = await call<{ targetType: number, targetId: number, type: string }, { active: boolean, total: number }>(
      '/sns.v1.ReactionService/ToggleReaction',
      { targetType: 1, targetId: postId, type: 'like' },
      headers
    )
    setPosts(prev => prev.map(p => p.id === postId ? { ...p, likedByMe: r.active, likeCount: r.total } : p))
  }

  return (
    <main>
      <h1 style={{ fontSize: 20, fontWeight: 'bold', marginBottom: 8 }}>Feed ({tenantSlug})</h1>
      <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        <input value={newBody} onChange={e => setNewBody(e.target.value)} placeholder="いまどうしてる？" style={{ flex: 1, padding: 8 }} />
        <button onClick={createPost} disabled={!newBody.trim()}>投稿</button>
      </div>
      <ul style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {posts.map(p => (
          <li key={p.id} style={{ border: '1px solid #ddd', padding: 12, borderRadius: 8 }}>
            <div style={{ fontSize: 12, color: '#666' }}>by {String(p.authorUserId)} at {new Date(p.createdAt).toLocaleString()}</div>
            <div style={{ margin: '8px 0' }}>{p.body}</div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button onClick={() => toggleLike(Number(p.id))}>{p.likedByMe ? '♥' : '♡'} {p.likeCount}</button>
              <span>💬 {p.commentCount}</span>
            </div>
          </li>
        ))}
      </ul>
    </main>
  )
}
