'use client'
import { useEffect, useState } from 'react'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'

type Comment = { id: number, postId: number, authorUserId: number, body: string, createdAt: string }

async function call<TReq extends object, TRes>(path: string, req: TReq, headers: Record<string,string>): Promise<TRes> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Accept': 'application/json', 'Connect-Protocol-Version': '1', ...headers },
    body: JSON.stringify(req ?? {}),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<TRes>
}

export default function PostDetail({ params }: { params: { id: string } }) {
  const postId = Number(params.id)
  const headers = { 'X-Tenant': 'acme', 'X-User': 'u_alice' }
  const [comments, setComments] = useState<Comment[]>([])
  const [body, setBody] = useState('')

  useEffect(() => {
    ;(async () => {
      await call('/sns.v1.TenantService/GetMe', {}, headers)
      const r = await call<{ post_id: number }, { items: Comment[] }>(
        '/sns.v1.TimelineService/ListComments',
        { post_id: postId } as any,
        headers
      )
      setComments(r.items)
    })()
  }, [postId])

  const submit = async () => {
    const r = await call<{ post_id: number, body: string }, { comment: Comment }>(
      '/sns.v1.TimelineService/CreateComment',
      { post_id: postId, body },
      headers
    )
    setComments([...comments, r.comment])
    setBody('')
  }

  return (
    <main>
      <h1>Post #{postId}</h1>
      <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
        <input value={body} onChange={e => setBody(e.target.value)} placeholder="コメントを書く" style={{ flex: 1, padding: 8 }} />
        <button onClick={submit} disabled={!body.trim()}>送信</button>
      </div>
      <ul style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {comments.map(c => (
          <li key={c.id} style={{ border: '1px solid #ddd', padding: 8 }}>
            <div style={{ fontSize: 12, color: '#666' }}>by {c.authorUserId} at {new Date(c.createdAt).toLocaleString()}</div>
            <div>{c.body}</div>
          </li>
        ))}
      </ul>
    </main>
  )
}
