import { useEffect, useState } from 'react'
import './App.css'

const API_URL = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')

const starterPoll = [
  { label: 'Morning coffee', votes: 42, color: 'coral' },
  { label: 'Iced matcha', votes: 28, color: 'mint' },
  { label: 'Chai latte', votes: 18, color: 'gold' },
]

async function requestApi(path, options = {}) {
  const token = localStorage.getItem('pulsevote_token')
  let response
  try {
    response = await fetch(`${API_URL}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    })
  } catch {
    throw new Error('Unable to reach the API. Check VITE_API_URL and your backend deployment.')
  }

  const contentType = response.headers.get('content-type') || ''
  const body = contentType.includes('application/json') ? await response.json() : {}
  if (!response.ok) {
    const error = new Error(body.message || `API request failed (${response.status})`)
    error.status = response.status
    throw error
  }
  return body.data
}

function navigate(path) {
  window.history.pushState({}, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

function toLocalDateTimeValue(date) {
  const pad = (value) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function getVoterStorageKey(pollId) {
  const signedInUser = JSON.parse(localStorage.getItem('pulsevote_user') || 'null')
  if (signedInUser?.id) return `votiva_voter_${pollId}_user_${signedInUser.id}`
  return `votiva_voter_${pollId}`
}

function getVoteStateKey(pollId) {
  const signedInUser = JSON.parse(localStorage.getItem('pulsevote_user') || 'null')
  if (signedInUser?.id) return `votiva_voted_${pollId}_user_${signedInUser.id}`
  return `votiva_voted_${pollId}`
}

async function copyPollLink(shareCode) {
  const link = `${window.location.origin}/poll/${shareCode}`
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(link)
  } else {
    const input = document.createElement('input')
    input.value = link
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    input.remove()
  }
  return link
}

async function sharePoll(shareCode) {
  const link = `${window.location.origin}/poll/${shareCode}`
  if (navigator.share) {
    await navigator.share({ title: 'Vote on this VibeVote poll', url: link })
  } else {
    await copyPollLink(shareCode)
  }
}

function App() {
  const [path, setPath] = useState(window.location.pathname)
  const [backendStatus, setBackendStatus] = useState('checking')
  const [user, setUser] = useState(() => JSON.parse(localStorage.getItem('pulsevote_user') || 'null'))

  useEffect(() => {
    const handleNavigation = () => setPath(window.location.pathname)
    window.addEventListener('popstate', handleNavigation)
    requestApi('/api/health').then(() => setBackendStatus('connected')).catch(() => setBackendStatus('offline'))
    return () => window.removeEventListener('popstate', handleNavigation)
  }, [])

  function handleAuth(authData) {
    localStorage.setItem('pulsevote_token', authData.token)
    localStorage.setItem('pulsevote_user', JSON.stringify(authData.user))
    setUser(authData.user)
    navigate('/dashboard')
  }

  function logout() {
    localStorage.removeItem('pulsevote_token')
    localStorage.removeItem('pulsevote_user')
    setUser(null)
    navigate('/')
  }

  const publicMatch = path.match(/^\/poll\/([^/]+)$/)
  let page = <Home backendStatus={backendStatus} />
  if (path === '/login' || path === '/signup') page = <AuthPage mode={path.slice(1)} onSuccess={handleAuth} />
  if (path === '/dashboard' || path === '/admin') page = user ? <Dashboard user={user} onLogout={logout} /> : <AuthPage mode="login" onSuccess={handleAuth} />
  if (path === '/create') page = user ? <CreatePoll /> : <AuthPage mode="login" onSuccess={handleAuth} />
  const editMatch = path.match(/^\/edit\/([^/]+)$/)
  if (editMatch) page = user ? <EditPoll pollId={editMatch[1]} /> : <AuthPage mode="login" onSuccess={handleAuth} />
  if (publicMatch) page = <PublicPoll shareCode={publicMatch[1]} />

  return <><Header user={user} onLogout={logout} backendStatus={backendStatus} />{page}</>
}

function Header({ user, onLogout, backendStatus }) {
  return <nav className="topbar">
    <a className="brand" href="/" onClick={(event) => { event.preventDefault(); navigate('/') }} aria-label="VibeVote home"><span className="brand-mark">◒</span><span>Vibe<span>Vote</span></span></a>
    <div className="nav-links"><a href="/" onClick={(event) => { event.preventDefault(); navigate('/') }}>Explore</a><a href="/#how">How it works</a></div>
    <div className="nav-actions">
      <span className={`connection ${backendStatus}`}><span className="live-dot" /> API {backendStatus}</span>
      {user ? <><button className="text-button" onClick={() => navigate('/dashboard')}>Dashboard</button><button className="dark-button" onClick={onLogout}>Log out</button></> : <><button className="text-button" onClick={() => navigate('/login')}>Log in</button><button className="dark-button" onClick={() => navigate('/create')}>Create a poll <span>↗</span></button></>}
    </div>
  </nav>
}

function Home({ backendStatus }) {
  const [poll, setPoll] = useState(starterPoll)
  const [selected, setSelected] = useState(null)
  const total = poll.reduce((sum, option) => sum + option.votes, 0)
  function castVote(index) {
    setSelected(index)
    setPoll((current) => current.map((option, optionIndex) => optionIndex === index ? { ...option, votes: option.votes + 1 } : option))
  }
  return <main className="app-shell">
    <section className="hero-grid">
      <div className="hero-copy"><p className="eyebrow"><span className="live-dot" /> LIVE OPINIONS, RIGHT NOW · API {backendStatus.toUpperCase()}</p><h1>Ask the room.<br /><em>Feel the pulse.</em></h1><p className="lede">Create a poll, share the link, and watch the room move. Your account and every poll are now powered by the VibeVote API.</p><div className="hero-ctas"><button className="coral-button" onClick={() => navigate('/create')}>Start a new poll <span>→</span></button><button className="play-button" onClick={() => navigate('/login')}>▷ <span>Sign in to your account</span></button></div><div className="proof"><div className="avatar-stack"><i>R</i><i>M</i><i>K</i><i>+</i></div><span><strong>One shared space</strong><br />for every point of view</span></div></div>
      <div className="poll-stage" id="explore"><div className="stage-note note-top">03 <span>people are<br />already here</span></div><div className="poll-card"><div className="poll-meta"><span className="status-pill"><span className="live-dot" /> LIVE NOW</span><span>Example poll</span></div><h2>What’s powering your<br />morning today?</h2><p className="poll-sub">One choice. Zero overthinking.</p><div className="options">{poll.map((option, index) => { const percent = Math.round((option.votes / total) * 100); return <button className={`poll-option ${selected === index ? 'selected' : ''}`} key={option.label} onClick={() => castVote(index)}><span className={`option-icon ${option.color}`}>{index === 0 ? '☕' : index === 1 ? '✦' : '◉'}</span><span className="option-label">{option.label}</span><span className="option-percent">{percent}%</span><span className={`bar ${option.color}`} style={{ width: `${percent}%` }} /></button> })}</div><div className="poll-footer"><span><span className="pulse-bars">▂▅▇</span> {total} votes · live demo</span><button className="share-button" onClick={() => navigate('/login')}>Create yours ↗</button></div></div><div className="stage-note note-bottom"><span className="sparkle">✦</span> no account<br />needed to vote</div></div>
    </section><section className="ticker"><span>WHAT PEOPLE ARE ASKING</span><b>⋆</b><span>TEAM DECISIONS</span><b>⋆</b><span>QUICK CHECK-INS</span><b>⋆</b><span>HOT TAKES</span></section><section className="bottom-section" id="how"><p className="eyebrow">BUILT FOR THE MOMENT</p><h2>Small question.<br /><em>Big signal.</em></h2><p>Sign up to create real polls, share public links, and collect votes through the connected backend.</p></section>
  </main>
}

function AuthPage({ mode, onSuccess }) {
  const isSignup = mode === 'signup'
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  async function submit(event) {
    event.preventDefault(); setBusy(true); setError('')
    try { const data = await requestApi(`/api/auth/${isSignup ? 'signup' : 'login'}`, { method: 'POST', body: JSON.stringify(form) }); onSuccess(data) } catch (requestError) { setError(requestError.message) } finally { setBusy(false) }
  }
  return <main className="form-page"><form className="form-card" onSubmit={submit}><p className="eyebrow">VIBEVOTE ACCOUNT</p><h1>{isSignup ? 'Join the room.' : 'Welcome back.'}</h1><p className="form-intro">{isSignup ? 'Create an account and start collecting real opinions.' : 'Sign in to manage your polls.'}</p>{isSignup && <label>Name<input required minLength="2" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></label>}<label>Email<input required type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} /></label><label>Password<input required minLength="6" type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} /></label>{error && <p className="error-message">{error}</p>}<button className="coral-button form-submit" disabled={busy}>{busy ? 'Connecting...' : isSignup ? 'Create account →' : 'Log in →'}</button><button type="button" className="form-switch" onClick={() => navigate(isSignup ? '/login' : '/signup')}>{isSignup ? 'Already have an account? Log in' : 'New here? Create an account'}</button></form></main>
}

function Dashboard({ user, onLogout }) {
  const [polls, setPolls] = useState([])
  const [stats, setStats] = useState({})
  const [error, setError] = useState('')
  const [copied, setCopied] = useState('')

  useEffect(() => {
    let active = true
    async function loadDashboard() {
      try {
        const data = await requestApi('/api/polls')
        if (!active) return
        setPolls(data.polls || [])
        const entries = await Promise.all((data.polls || []).map(async (poll) => [poll.id, await requestApi(`/api/polls/${poll.id}/stats`)]))
        if (active) setStats(Object.fromEntries(entries))
      } catch (requestError) {
        if (active) setError(requestError.message)
      }
    }
    loadDashboard()
    const timer = window.setInterval(loadDashboard, 3000)
    return () => { active = false; window.clearInterval(timer) }
  }, [])

  async function copyLink(shareCode) {
    await copyPollLink(shareCode)
    setCopied(shareCode)
    window.setTimeout(() => setCopied(''), 1800)
  }

  async function deletePoll(pollId) {
    if (!window.confirm('Are you sure you want to delete this poll?')) return
    try {
      await requestApi(`/api/polls/${pollId}`, { method: 'DELETE' })
      setPolls((current) => current.filter((poll) => poll.id !== pollId))
    } catch (requestError) {
      setError(requestError.message)
    }
  }

  return <main className="workspace"><div className="workspace-heading"><div><p className="eyebrow">ADMIN CONTROL ROOM · LIVE</p><h1>Hello, {user.name}.</h1><p>Track every response as it arrives. Scores refresh automatically.</p></div><button className="coral-button" onClick={() => navigate('/create')}>New poll <span>→</span></button></div>{error && <p className="error-message">{error}</p>}<div className="poll-list">{polls.length ? polls.map((poll) => <PollAdminCard key={poll.id} poll={poll} stats={stats[poll.id]} copied={copied === poll.shareCode} onCopy={() => copyLink(poll.shareCode)} onDelete={() => deletePoll(poll.id)} />) : <div className="empty-state"><h2>Your first question is waiting.</h2><p>Create a poll and share it with your room.</p><button className="dark-button" onClick={() => navigate('/create')}>Create a poll</button></div>}</div><button className="text-button" onClick={onLogout}>Log out</button></main>
}

function PollAdminCard({ poll, stats, copied, onCopy, onDelete }) {
  const totalVotes = stats?.totalVotes || 0
  return <article className="admin-card"><div className="admin-card-heading"><div><span className={poll.isActive ? 'active-label' : 'closed-label'}>{poll.isActive ? '● LIVE' : 'CLOSED'}</span><h2>{poll.question}</h2><p>{stats ? `Updated ${new Date(stats.updatedAt).toLocaleTimeString()}` : 'Loading live score...'}</p></div><strong className="vote-total"><span>{totalVotes}</span> votes</strong></div><div className="score-list">{(stats?.optionCounts || poll.options.map((option) => ({ optionId: option.id, label: option.text, count: 0, percentage: 0 }))).map((option) => <div className="score-row" key={option.optionId}><div className="score-label"><span>{option.label}</span><strong>{option.count} · {Math.round(option.percentage)}%</strong></div><div className="score-track"><span style={{ width: `${option.percentage}%` }} /></div></div>)}</div><div className="admin-card-actions"><button className="icon-action" title="Open voting page" onClick={() => navigate(`/poll/${poll.shareCode}`)}>↗ <span>Open</span></button><button className="icon-action" title="Share poll" onClick={() => sharePoll(poll.shareCode)}>⌯ <span>Share</span></button><button className="icon-action" title="Copy voting link" onClick={onCopy}>▣ <span>{copied ? 'Copied' : 'Copy link'}</span></button><button className="icon-action" title="Edit poll" onClick={() => navigate(`/edit/${poll.id}`)}>✎ <span>Edit</span></button><button className="icon-action danger-action" title="Delete poll" onClick={onDelete}>× <span>Delete</span></button><span className="share-url">/poll/{poll.shareCode}</span></div></article>
}

function CreatePoll() {
  const [question, setQuestion] = useState(''); const [options, setOptions] = useState(['', '']); const [expiresAt, setExpiresAt] = useState(''); const [error, setError] = useState(''); const [busy, setBusy] = useState(false)
  function setExpiry(hours) { setExpiresAt(toLocalDateTimeValue(new Date(Date.now() + hours * 60 * 60 * 1000))) }
  async function submit(event) { event.preventDefault(); setBusy(true); setError(''); try { await requestApi('/api/polls', { method: 'POST', body: JSON.stringify({ question, options: options.filter(Boolean).map((text) => ({ text, vibeMessage: '' })), allowAnonymousVoting: true, allowOneVotePerBrowser: true, showPercentages: true, showTotalVoteCount: true, expiresAt: expiresAt ? new Date(expiresAt).toISOString() : null }) }); navigate(`/admin`) } catch (requestError) { setError(requestError.message) } finally { setBusy(false) } }
  return <main className="form-page"><form className="form-card wide-form" onSubmit={submit}><p className="eyebrow">NEW POLL · LIVE TRACKING ENABLED</p><h1>Ask the room.</h1><p className="form-intro">Write a clear question, add answer choices, then share the voting link from your admin control room.</p><label>Question<input required minLength="6" maxLength="200" value={question} onChange={(event) => setQuestion(event.target.value)} placeholder="What should we do next?" /></label><div className="option-fields"><span className="field-title">Answer options</span>{options.map((option, index) => <input required key={index} value={option} onChange={(event) => setOptions(options.map((current, optionIndex) => optionIndex === index ? event.target.value : current))} placeholder={`Option ${index + 1}`} />)}</div>{options.length < 10 && <button type="button" className="add-option" onClick={() => setOptions([...options, ''])}>+ Add another option</button>}<div className="expiry-field"><label>Close voting <input type="datetime-local" min={new Date().toISOString().slice(0, 16)} value={expiresAt} onChange={(event) => setExpiresAt(event.target.value)} /></label><div className="expiry-presets"><button type="button" onClick={() => setExpiry(1)}>1 hour</button><button type="button" onClick={() => setExpiry(24)}>24 hours</button><button type="button" onClick={() => setExpiry(72)}>3 days</button><button type="button" onClick={() => setExpiresAt('')}>No expiry</button></div></div>{error && <p className="error-message">{error}</p>}<button className="coral-button form-submit" disabled={busy}>{busy ? 'Creating...' : 'Create poll →'}</button></form></main>
}

function EditPoll({ pollId }) {
  const [poll, setPoll] = useState(null)
  const [question, setQuestion] = useState('')
  const [options, setOptions] = useState([])
  const [expiresAt, setExpiresAt] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    requestApi('/api/polls').then((data) => {
      const found = (data.polls || []).find((item) => item.id === pollId)
      if (!found) throw new Error('Poll not found')
      setPoll(found)
      setQuestion(found.question)
      setOptions(found.options)
      setExpiresAt(found.expiresAt ? new Date(found.expiresAt).toISOString().slice(0, 16) : '')
    }).catch((requestError) => setError(requestError.message))
  }, [pollId])

  async function submit(event) {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      await requestApi(`/api/polls/${pollId}`, { method: 'PUT', body: JSON.stringify({ question, options: options.filter((option) => option.text.trim()).map((option) => ({ id: option.id, text: option.text, vibeMessage: option.vibeMessage })), allowAnonymousVoting: poll.allowAnonymousVoting, allowOneVotePerBrowser: poll.allowOneVotePerBrowser, showPercentages: poll.showPercentages, showTotalVoteCount: poll.showTotalVoteCount, expiresAt: expiresAt ? new Date(expiresAt).toISOString() : null }) })
      navigate('/admin')
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setBusy(false)
    }
  }

  if (!poll && !error) return <main className="form-page"><p className="form-intro">Loading poll...</p></main>
  return <main className="form-page"><form className="form-card wide-form" onSubmit={submit}><p className="eyebrow">EDIT POLL</p><h1>Refine the question.</h1><label>Question<input required minLength="6" maxLength="200" value={question} onChange={(event) => setQuestion(event.target.value)} /></label><div className="option-fields"><span className="field-title">Answer options and Vote Vibe</span>{options.map((option, index) => <div className="edit-option" key={option.id || index}><input required value={option.text} onChange={(event) => setOptions(options.map((current, optionIndex) => optionIndex === index ? { ...current, text: event.target.value } : current))} /><input value={option.vibeMessage || ''} placeholder="Vote Vibe message" onChange={(event) => setOptions(options.map((current, optionIndex) => optionIndex === index ? { ...current, vibeMessage: event.target.value } : current))} /></div>)}</div><label>Close voting <input type="datetime-local" value={expiresAt} onChange={(event) => setExpiresAt(event.target.value)} /></label>{error && <p className="error-message">{error}</p>}<button className="coral-button form-submit" disabled={busy}>{busy ? 'Saving...' : 'Save changes →'}</button></form></main>
}

function PublicPoll({ shareCode }) {
  const [poll, setPoll] = useState(null); const [selectedOptionId, setSelectedOptionId] = useState(null); const [hasVoted, setHasVoted] = useState(false); const [voteVibe, setVoteVibe] = useState(null); const [message, setMessage] = useState('Loading poll...'); const [error, setError] = useState(''); const [busy, setBusy] = useState(false); const [copied, setCopied] = useState(false); const [now, setNow] = useState(0)
  useEffect(() => { requestApi(`/api/polls/public/${shareCode}`).then((data) => { setPoll(data.poll); const savedVote = JSON.parse(localStorage.getItem(getVoteStateKey(data.poll.id)) || 'null'); if (savedVote?.voted) { setHasVoted(true); setSelectedOptionId(savedVote.optionId || null); setVoteVibe(savedVote.voteVibe || null); } setMessage('') }).catch((requestError) => { setError(requestError.message); setMessage('') }) }, [shareCode])
  useEffect(() => { const timer = window.setInterval(() => setNow(Date.now()), 1000); return () => window.clearInterval(timer) }, [])
  const expired = poll?.expiresAt && new Date(poll.expiresAt).getTime() <= now
  async function vote() { const selectedOption = poll?.options.find((option) => option.id === selectedOptionId); if (!poll?.id || !selectedOption || hasVoted || busy || expired) return; setBusy(true); setError(''); setMessage('Recording your vote...'); try { const voterStorageKey = getVoterStorageKey(poll.id); let voterKey = localStorage.getItem(voterStorageKey); if (!voterKey) { voterKey = crypto.randomUUID(); localStorage.setItem(voterStorageKey, voterKey) } const data = await requestApi(`/api/polls/${poll.id}/vote`, { method: 'POST', body: JSON.stringify({ optionId: selectedOption.id, voterKey }) }); const savedVote = { pollId: poll.id, optionId: selectedOption.id, voted: true, voteVibe: data.vibeMessage || selectedOption.vibeMessage || null }; localStorage.setItem(getVoteStateKey(poll.id), JSON.stringify(savedVote)); setHasVoted(true); setVoteVibe(savedVote.voteVibe); setMessage('') } catch (requestError) { if (requestError.status === 409) { setHasVoted(true); setMessage(''); setError('You have already voted in this poll.') } else { setError(requestError.message); setMessage('') } } finally { setBusy(false) } }
  async function copyLink() { await copyPollLink(shareCode); setCopied(true); window.setTimeout(() => setCopied(false), 1800) }
  const timeLeft = poll?.expiresAt ? formatTimeLeft(new Date(poll.expiresAt).getTime() - now) : ''
  return <main className="public-page"><div className="public-card"><p className="eyebrow">PUBLIC POLL · {shareCode}</p>{poll?.expiresAt && <div className={`expiry-banner ${expired ? 'expired' : ''}`}>{expired ? 'Voting has closed' : `Voting closes in ${timeLeft}`}</div>}{message && <p className={`form-intro ${message.startsWith('Recording') ? 'pending-message' : ''}`}>{message}</p>}{error && <p className="error-message">{error}</p>}{voteVibe && <p className="success-message">✨ {voteVibe}</p>}{hasVoted && <p className="already-voted">✓ You have already voted in this poll.</p>}{poll && <><h1>{poll.question}</h1>{poll.description && <p className="form-intro">{poll.description}</p>}{!hasVoted && <div className="public-options">{poll.options.map((option) => <label className={`public-option ${selectedOptionId === option.id ? 'selected' : ''}`} key={option.id}><input type="radio" name="poll-option" value={option.id} checked={selectedOptionId === option.id} onChange={() => setSelectedOptionId(option.id)} />{option.text}<span className="choice-check">{selectedOptionId === option.id ? '✓' : ''}</span></label>)}</div>}{!hasVoted && <button className="coral-button form-submit vote-submit" disabled={!selectedOptionId || busy || !poll.isActive || expired} onClick={vote}>{busy ? 'Submitting...' : expired || !poll.isActive ? 'Voting closed' : selectedOptionId ? 'Vote' : 'Choose an option first'}</button>}<div className="public-share"><button className="icon-action" onClick={() => sharePoll(shareCode)}>⌯ <span>Share poll</span></button><button className="icon-action" onClick={copyLink}>▣ <span>{copied ? 'Link copied' : 'Copy voting link'}</span></button></div></>}</div></main>
}

function formatTimeLeft(milliseconds) {
  if (milliseconds <= 0) return '0m'
  const totalMinutes = Math.floor(milliseconds / 60000)
  const days = Math.floor(totalMinutes / 1440)
  const hours = Math.floor((totalMinutes % 1440) / 60)
  const minutes = totalMinutes % 60
  return days ? `${days}d ${hours}h` : hours ? `${hours}h ${minutes}m` : `${minutes}m`
}

export default App
