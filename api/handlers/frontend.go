package handlers

import "github.com/gofiber/fiber/v2"

func Frontend(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(frontendHTML)
}

const frontendHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>Document Processing Platform</title>
  <link rel="preconnect" href="https://fonts.googleapis.com"/>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;600;700&family=Syne:wght@400;700;800&display=swap" rel="stylesheet"/>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    :root {
      --bg:       #06080f;
      --surface:  #0d1120;
      --border:   #1e2840;
      --accent:   #00e5a0;
      --accent2:  #0066ff;
      --danger:   #ff4560;
      --text:     #e8eaf0;
      --muted:    #5a6282;
      --mono:     'JetBrains Mono', monospace;
      --sans:     'Syne', sans-serif;
    }
    html, body { height: 100%; background: var(--bg); color: var(--text); font-family: var(--mono); overflow-x: hidden; }
    body::before {
      content: ''; position: fixed; inset: 0;
      background-image: linear-gradient(rgba(0,229,160,.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0,229,160,.03) 1px, transparent 1px);
      background-size: 40px 40px; pointer-events: none; z-index: 0;
    }
    .blob { position: fixed; border-radius: 50%; filter: blur(120px); opacity: .18; pointer-events: none; z-index: 0; }
    .blob-1 { width: 500px; height: 500px; background: var(--accent); top: -200px; left: -100px; }
    .blob-2 { width: 400px; height: 400px; background: var(--accent2); bottom: -150px; right: -100px; }
    .shell { position: relative; z-index: 1; min-height: 100vh; display: flex; flex-direction: column; align-items: center; padding: 60px 20px 80px; }
    header { text-align: center; margin-bottom: 48px; }
    .badge { display: inline-flex; align-items: center; gap: 8px; background: rgba(0,229,160,.08); border: 1px solid rgba(0,229,160,.25); border-radius: 100px; padding: 6px 14px; font-size: 11px; letter-spacing: .12em; text-transform: uppercase; color: var(--accent); margin-bottom: 20px; }
    .badge span { width: 6px; height: 6px; background: var(--accent); border-radius: 50%; animation: pulse 2s infinite; }
    @keyframes pulse { 0%,100%{opacity:1;transform:scale(1)} 50%{opacity:.4;transform:scale(1.4)} }
    h1 { font-family: var(--sans); font-size: clamp(2.2rem, 5.5vw, 3.9rem); font-weight: 800; line-height: 1.05; letter-spacing: -.02em; }
    h1 em { font-style: normal; color: var(--accent); }
    .sub { margin-top: 14px; color: var(--muted); font-size: 13px; letter-spacing: .04em; }
    .card { width: 100%; max-width: 680px; background: var(--surface); border: 1px solid var(--border); border-radius: 16px; padding: 32px; position: relative; overflow: hidden; }
    .card::before { content: ''; position: absolute; inset: 0; background: linear-gradient(135deg, rgba(0,229,160,.04) 0%, transparent 60%); pointer-events: none; }
    .field { position: relative; z-index: 1; margin-bottom: 18px; }
    .field label { display: block; font-size: 11px; color: var(--muted); margin-bottom: 7px; letter-spacing: .1em; text-transform: uppercase; }
    .field input { width: 100%; padding: 11px 14px; background: var(--bg); border: 1px solid var(--border); border-radius: 9px; color: var(--text); font-family: var(--mono); font-size: 13px; transition: border-color .2s; }
    .field input::placeholder { color: #39425e; }
    .field input:focus { outline: none; border-color: var(--accent); }
    .seg { display: flex; gap: 6px; background: var(--bg); border: 1px solid var(--border); border-radius: 11px; padding: 5px; margin-bottom: 18px; position: relative; z-index: 1; }
    .seg button { flex: 1; padding: 10px; border: none; border-radius: 7px; background: transparent; color: var(--muted); font-family: var(--mono); font-size: 12px; cursor: pointer; transition: background .2s, color .2s; }
    .seg button:hover { color: var(--text); }
    .seg button.active { background: var(--accent); color: #000; font-weight: 600; }
    #drop-zone { border: 2px dashed var(--border); border-radius: 12px; padding: 40px 24px; display: flex; flex-direction: column; align-items: center; gap: 12px; cursor: pointer; transition: border-color .2s, background .2s; text-align: center; position: relative; z-index: 1; }
    #drop-zone:hover, #drop-zone.drag-over { border-color: var(--accent); background: rgba(0,229,160,.04); }
    #drop-zone input[type=file] { position: absolute; inset: 0; opacity: 0; cursor: pointer; }
    .drop-icon { width: 52px; height: 52px; background: rgba(0,229,160,.1); border: 1px solid rgba(0,229,160,.2); border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 22px; transition: transform .2s; }
    #drop-zone:hover .drop-icon { transform: translateY(-3px); }
    .drop-title { font-family: var(--sans); font-size: 16px; font-weight: 700; }
    .drop-sub { font-size: 12px; color: var(--muted); }
    .drop-meta { font-size: 11px; color: var(--muted); margin-top: 4px; }
    #file-chip { display: none; align-items: center; gap: 10px; background: rgba(0,229,160,.07); border: 1px solid rgba(0,229,160,.2); border-radius: 8px; padding: 10px 14px; font-size: 13px; margin-top: 16px; position: relative; z-index: 1; }
    #file-chip.show { display: flex; }
    .chip-name { flex: 1; color: var(--accent); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .chip-type { font-size: 10px; color: var(--muted); border: 1px solid var(--border); border-radius: 4px; padding: 2px 7px; text-transform: uppercase; letter-spacing: .06em; }
    .chip-remove { background: none; border: none; color: var(--muted); cursor: pointer; font-size: 16px; line-height: 1; padding: 0; transition: color .15s; }
    .chip-remove:hover { color: var(--danger); }
    details.webhook { margin-top: 16px; border: 1px solid var(--border); border-radius: 10px; overflow: hidden; position: relative; z-index: 1; }
    details.webhook summary { padding: 12px 16px; cursor: pointer; font-size: 12px; color: var(--muted); list-style: none; display: flex; align-items: center; gap: 8px; user-select: none; transition: color .2s; }
    details.webhook summary:hover { color: var(--text); }
    details.webhook summary::after { content: '+'; margin-left: auto; font-size: 16px; transition: transform .2s; }
    details.webhook[open] summary::after { transform: rotate(45deg); }
    .webhook-body { padding: 0 16px 16px; }
    .webhook-body input { width: 100%; padding: 11px 14px; background: var(--bg); border: 1px solid var(--border); border-radius: 9px; color: var(--text); font-family: var(--mono); font-size: 12px; }
    .webhook-body input:focus { outline: none; border-color: var(--accent); }
    .webhook-hint { font-size: 10px; color: var(--muted); margin-top: 8px; line-height: 1.5; }
    #upload-btn { width: 100%; margin-top: 20px; padding: 14px; border: none; border-radius: 10px; background: var(--accent); color: #000; font-family: var(--sans); font-size: 15px; font-weight: 700; letter-spacing: .04em; cursor: pointer; transition: opacity .2s, transform .15s; display: flex; align-items: center; justify-content: center; gap: 8px; position: relative; z-index: 1; }
    #upload-btn:disabled { opacity: .45; cursor: not-allowed; transform: none !important; }
    #upload-btn:not(:disabled):hover { transform: translateY(-1px); opacity: .92; }
    .pipe-head { display: none; align-items: center; justify-content: space-between; margin-top: 28px; margin-bottom: 4px; position: relative; z-index: 1; }
    .pipe-head.show { display: flex; }
    .pipe-title { font-family: var(--sans); font-size: 12px; letter-spacing: .1em; text-transform: uppercase; color: var(--muted); }
    .live-badge { display: none; align-items: center; gap: 6px; font-size: 11px; letter-spacing: .06em; padding: 4px 10px; border-radius: 100px; }
    .live-badge.show { display: inline-flex; }
    .live-badge.sse { color: var(--accent); background: rgba(0,229,160,.08); border: 1px solid rgba(0,229,160,.25); }
    .live-badge.poll { color: #ff9600; background: rgba(255,150,0,.08); border: 1px solid rgba(255,150,0,.25); }
    .live-dot { width: 7px; height: 7px; border-radius: 50%; background: currentColor; animation: pulse 1.3s infinite; box-shadow: 0 0 8px currentColor; }
    #pipeline { display: none; margin-top: 14px; flex-direction: column; position: relative; z-index: 1; }
    #pipeline.show { display: flex; }
    .pipe-step { display: flex; align-items: flex-start; gap: 16px; position: relative; padding-bottom: 22px; }
    .pipe-step:last-child { padding-bottom: 0; }
    .pipe-step:not(:last-child)::after { content: ''; position: absolute; left: 16px; top: 34px; bottom: 0; width: 1px; background: var(--border); }
    .pipe-step.done::after { background: var(--accent); opacity: .4; }
    .step-dot { width: 33px; height: 33px; border-radius: 50%; border: 1.5px solid var(--border); background: var(--bg); display: flex; align-items: center; justify-content: center; font-size: 13px; flex-shrink: 0; transition: all .3s; position: relative; z-index: 1; }
    .pipe-step.active .step-dot { border-color: var(--accent); background: rgba(0,229,160,.12); box-shadow: 0 0 0 4px rgba(0,229,160,.08); }
    .pipe-step.done .step-dot { border-color: var(--accent); background: var(--accent); color: #000; }
    .pipe-step.failed .step-dot { border-color: var(--danger); background: rgba(255,69,96,.12); }
    .dot-spinner { width: 14px; height: 14px; border: 2px solid rgba(0,229,160,.3); border-top-color: var(--accent); border-radius: 50%; animation: spin .8s linear infinite; display: none; }
    .pipe-step.active .dot-spinner { display: block; }
    .pipe-step.active .dot-icon { display: none; }
    @keyframes spin { to { transform: rotate(360deg); } }
    .step-body { flex: 1; padding-top: 4px; }
    .step-label { font-family: var(--sans); font-size: 14px; font-weight: 700; transition: color .3s; }
    .pipe-step.active .step-label { color: var(--accent); }
    .pipe-step.done .step-label { color: var(--accent); }
    .pipe-step.failed .step-label { color: var(--danger); }
    .step-desc { font-size: 11px; color: var(--muted); margin-top: 3px; line-height: 1.5; }
    .step-tag { display: inline-block; margin-top: 6px; background: rgba(0,229,160,.06); border: 1px solid rgba(0,229,160,.15); border-radius: 4px; padding: 3px 8px; font-size: 10px; color: var(--muted); letter-spacing: .04em; word-break: break-all; }
    #download-block { display: none; margin-top: 28px; background: rgba(0,229,160,.06); border: 1px solid rgba(0,229,160,.2); border-radius: 12px; padding: 24px; flex-direction: column; align-items: center; gap: 14px; text-align: center; animation: fadeUp .4s ease; position: relative; z-index: 1; }
    @keyframes fadeUp { from{opacity:0;transform:translateY(10px)} to{opacity:1;transform:none} }
    #download-block.show { display: flex; }
    .dl-icon { font-size: 34px; animation: bounce .6s ease; }
    @keyframes bounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-8px)} }
    .dl-title { font-family: var(--sans); font-size: 18px; font-weight: 800; color: var(--accent); }
    .dl-sub { font-size: 11px; color: var(--muted); }
    .thumb-preview { opacity: 0; transform: translateY(10px) scale(.97); transition: opacity .55s ease, transform .55s ease; }
    .thumb-preview.show { opacity: 1; transform: none; }
    .thumb-preview img { width: 200px; height: auto; display: block; border-radius: 8px; border: 1px solid var(--border); box-shadow: 0 10px 40px rgba(0,0,0,.5); }
    .thumb-caption { font-size: 10px; color: var(--muted); margin-top: 8px; letter-spacing: .06em; }
    #dl-btn { display: inline-flex; align-items: center; gap: 8px; background: var(--accent); color: #000; font-family: var(--sans); font-weight: 700; font-size: 14px; border: none; border-radius: 8px; padding: 12px 28px; cursor: pointer; text-decoration: none; transition: opacity .2s, transform .15s; }
    #dl-btn:hover { opacity: .88; transform: translateY(-1px); }
    #new-btn { background: none; border: 1px solid var(--border); color: var(--muted); font-family: var(--mono); font-size: 12px; border-radius: 6px; padding: 8px 16px; cursor: pointer; transition: border-color .2s, color .2s; }
    #new-btn:hover { border-color: var(--accent); color: var(--accent); }
    .banner { display: none; margin-top: 20px; border-radius: 10px; padding: 14px 18px; font-size: 12px; gap: 8px; align-items: flex-start; position: relative; z-index: 1; line-height: 1.5; }
    .banner.show { display: flex; }
    #error-banner { background: rgba(255,69,96,.08); border: 1px solid rgba(255,69,96,.3); color: var(--danger); }
    #rate-banner { background: rgba(255,150,0,.08); border: 1px solid rgba(255,150,0,.3); color: #ff9600; }
    .info-panel { width: 100%; max-width: 680px; margin-top: 24px; }
    details.arch { background: var(--surface); border: 1px solid var(--border); border-radius: 12px; overflow: hidden; }
    details.arch summary { padding: 16px 20px; cursor: pointer; font-size: 12px; color: var(--muted); letter-spacing: .08em; text-transform: uppercase; list-style: none; display: flex; align-items: center; gap: 8px; user-select: none; transition: color .2s; }
    details.arch summary:hover { color: var(--text); }
    details.arch summary::after { content: '▾'; margin-left: auto; transition: transform .2s; }
    details.arch[open] summary::after { transform: rotate(180deg); }
    .arch-body { padding: 0 20px 20px; }
    .arch-group { font-size: 10px; color: var(--muted); letter-spacing: .14em; text-transform: uppercase; margin: 14px 0 6px; padding-left: 12px; }
    .arch-row { display: flex; align-items: center; gap: 12px; padding: 8px 12px; border-radius: 6px; font-size: 12px; }
    .arch-row:hover { background: rgba(255,255,255,.03); }
    .arch-num { width: 22px; height: 22px; background: rgba(0,229,160,.08); border: 1px solid rgba(0,229,160,.2); border-radius: 4px; display: flex; align-items: center; justify-content: center; font-size: 10px; color: var(--accent); flex-shrink: 0; }
    .arch-num.q { background: rgba(0,102,255,.1); border-color: rgba(0,102,255,.3); color: #4d94ff; }
    .arch-svc { color: var(--accent); font-weight: 600; min-width: 116px; }
    .arch-row.queue .arch-svc { color: #4d94ff; }
    .arch-desc { color: var(--muted); }
    footer { margin-top: 48px; font-size: 11px; color: var(--muted); text-align: center; letter-spacing: .06em; }
    footer span { color: var(--accent); }
  </style>
</head>
<body>
  <div class="blob blob-1"></div>
  <div class="blob blob-2"></div>
  <div class="shell">
    <header>
      <div class="badge"><span></span> AWS Event-Driven Platform</div>
      <h1>Document <em>Processing</em> Platform</h1>
      <p class="sub">Markdown &amp; DOCX &rarr; PDF &middot; thumbnails &middot; webhooks &middot; live SSE updates</p>
    </header>

    <div class="card">
      <div class="field">
        <label for="api-key">API Key &middot; sent as X-API-Key</label>
        <input type="password" id="api-key" placeholder="sk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" autocomplete="off" />
      </div>

      <div class="seg" id="type-seg">
        <button data-type="auto" class="active">Auto-detect</button>
        <button data-type="md">Markdown</button>
        <button data-type="docx">DOCX</button>
      </div>

      <div id="drop-zone">
        <input type="file" id="file-input" accept=".md,.docx" />
        <div class="drop-icon">📄</div>
        <div class="drop-title">Drop your document here</div>
        <div class="drop-sub">or click to browse</div>
        <div class="drop-meta" id="drop-meta">.md (max 5 MB) or .docx (max 20 MB)</div>
      </div>

      <div id="file-chip">
        <span>📎</span>
        <span class="chip-name" id="chip-name">file</span>
        <span class="chip-type" id="chip-type">md</span>
        <button class="chip-remove" id="chip-remove" title="Remove">✕</button>
      </div>

      <details class="webhook">
        <summary>🔔 Webhook URL (optional)</summary>
        <div class="webhook-body">
          <input type="url" id="webhook-url" placeholder="https://your-server.com/callbacks/docplatform" />
          <div class="webhook-hint">HTTPS only. On completion we POST an HMAC-SHA256 signed payload (X-Signature header) to this URL &mdash; no need to keep this tab open.</div>
        </div>
      </details>

      <button id="upload-btn" disabled>
        <span id="btn-text">Select a file to convert</span>
      </button>

      <div class="banner" id="error-banner">⚠️ <span id="error-msg"></span></div>
      <div class="banner" id="rate-banner">⏳ Rate limit hit — max 5 uploads/min. Please wait a moment.</div>

      <div class="pipe-head" id="pipe-head">
        <span class="pipe-title">Pipeline</span>
        <span class="live-badge sse" id="live-badge"><span class="live-dot"></span><span id="live-text">Live</span></span>
      </div>

      <div id="pipeline">
        <div class="pipe-step" id="step-upload">
          <div class="step-dot"><div class="dot-spinner"></div><div class="dot-icon">⬆</div></div>
          <div class="step-body">
            <div class="step-label">Uploading to S3</div>
            <div class="step-desc">Sending your document to Amazon S3</div>
          </div>
        </div>
        <div class="pipe-step" id="step-queue">
          <div class="step-dot"><div class="dot-spinner"></div><div class="dot-icon">⬡</div></div>
          <div class="step-body">
            <div class="step-label">Job Queued (SQS)</div>
            <div class="step-desc">Routed to the matching SQS queue</div>
            <div class="step-tag" id="file-id-tag"></div>
          </div>
        </div>
        <div class="pipe-step" id="step-worker">
          <div class="step-dot"><div class="dot-spinner"></div><div class="dot-icon">⚙</div></div>
          <div class="step-body">
            <div class="step-label">Worker Processing</div>
            <div class="step-desc" id="worker-desc">A converter worker is processing the job</div>
          </div>
        </div>
        <div class="pipe-step" id="step-thumb">
          <div class="step-dot"><div class="dot-spinner"></div><div class="dot-icon">🖼</div></div>
          <div class="step-body">
            <div class="step-label">Generating Thumbnail</div>
            <div class="step-desc">worker-thumb renders page 1 of the PDF to JPEG</div>
          </div>
        </div>
        <div class="pipe-step" id="step-done">
          <div class="step-dot"><div class="dot-spinner"></div><div class="dot-icon">✓</div></div>
          <div class="step-body">
            <div class="step-label">Complete</div>
            <div class="step-desc">PDF + thumbnail stored in S3 · presigned URLs issued</div>
          </div>
        </div>
      </div>

      <div id="download-block">
        <div class="dl-icon">🎉</div>
        <div class="dl-title">Conversion Complete!</div>
        <div class="thumb-preview" id="thumb-preview">
          <img id="thumb-img" alt="PDF page 1 preview" />
          <div class="thumb-caption">page 1 preview</div>
        </div>
        <div class="dl-sub">Presigned URLs expire in 10 minutes</div>
        <a id="dl-btn" href="#" target="_blank" rel="noopener">⬇ Download PDF</a>
        <button id="new-btn">↩ Convert another file</button>
      </div>
    </div>

    <div class="info-panel">
      <details class="arch">
        <summary>🏗 System Architecture</summary>
        <div class="arch-body">
          <div class="arch-group">Services (5)</div>
          <div class="arch-row"><div class="arch-num">1</div><div class="arch-svc">API</div><div class="arch-desc">Go + Fiber. Auth, quota, upload, type routing, SSE stream</div></div>
          <div class="arch-row"><div class="arch-num">2</div><div class="arch-svc">worker-md</div><div class="arch-desc">Markdown → HTML (goldmark) → PDF (wkhtmltopdf)</div></div>
          <div class="arch-row"><div class="arch-num">3</div><div class="arch-svc">worker-docx</div><div class="arch-desc">DOCX → PDF via LibreOffice headless</div></div>
          <div class="arch-row"><div class="arch-num">4</div><div class="arch-svc">worker-thumb</div><div class="arch-desc">PDF page 1 → JPEG (pdftoppm) · fires webhooks</div></div>
          <div class="arch-row"><div class="arch-num">5</div><div class="arch-svc">Notification poller</div><div class="arch-desc">Runs in API · drains the notification queue · pushes SSE</div></div>
          <div class="arch-group">SQS Queues (4)</div>
          <div class="arch-row queue"><div class="arch-num q">A</div><div class="arch-svc">md-queue</div><div class="arch-desc">Markdown jobs → worker-md (DLQ after 3 retries)</div></div>
          <div class="arch-row queue"><div class="arch-num q">B</div><div class="arch-svc">docx-queue</div><div class="arch-desc">DOCX jobs → worker-docx (longer visibility timeout)</div></div>
          <div class="arch-row queue"><div class="arch-num q">C</div><div class="arch-svc">thumb-queue</div><div class="arch-desc">Thumbnail jobs → worker-thumb after PDF is ready</div></div>
          <div class="arch-row queue"><div class="arch-num q">D</div><div class="arch-svc">notification-queue</div><div class="arch-desc">Terminal results → API poller → browser via SSE</div></div>
          <div class="arch-group">Data Stores</div>
          <div class="arch-row"><div class="arch-num">S3</div><div class="arch-svc">S3</div><div class="arch-desc">uploads/ · pdfs/ · thumbnails/ — private, presigned access</div></div>
          <div class="arch-row"><div class="arch-num">DB</div><div class="arch-svc">DynamoDB</div><div class="arch-desc">jobs (status, urls) + tenants (api keys, quota)</div></div>
        </div>
      </details>
    </div>

    <footer>Document Processing Platform · Built with <span>Go · Fiber · Docker · Terraform · Ansible · Jenkins</span> · AWS</footer>
  </div>

  <script>
    var selectedFile = null;
    var currentFileId = null;
    var currentType = null;     // 'md' | 'docx' once an upload starts
    var typeMode = 'auto';      // 'auto' | 'md' | 'docx'
    var apiKey = '';
    var es = null;
    var sseOpen = false;
    var sseFallbackTimer = null;
    var pollInterval = null;
    var backstopInterval = null;
    var done = false;

    var apiKeyInput = document.getElementById('api-key');
    var dropZone    = document.getElementById('drop-zone');
    var dropMeta    = document.getElementById('drop-meta');
    var fileInput   = document.getElementById('file-input');
    var fileChip    = document.getElementById('file-chip');
    var chipName    = document.getElementById('chip-name');
    var chipType    = document.getElementById('chip-type');
    var chipRemove  = document.getElementById('chip-remove');
    var webhookInput = document.getElementById('webhook-url');
    var uploadBtn   = document.getElementById('upload-btn');
    var btnText     = document.getElementById('btn-text');
    var pipeHead    = document.getElementById('pipe-head');
    var pipeline    = document.getElementById('pipeline');
    var liveBadge   = document.getElementById('live-badge');
    var liveText    = document.getElementById('live-text');
    var dlBlock     = document.getElementById('download-block');
    var dlBtn       = document.getElementById('dl-btn');
    var newBtn      = document.getElementById('new-btn');
    var errorBanner = document.getElementById('error-banner');
    var errorMsg    = document.getElementById('error-msg');
    var rateBanner  = document.getElementById('rate-banner');
    var fileIdTag   = document.getElementById('file-id-tag');
    var workerDesc  = document.getElementById('worker-desc');
    var thumbPreview = document.getElementById('thumb-preview');
    var thumbImg    = document.getElementById('thumb-img');
    var steps = {
      upload: document.getElementById('step-upload'),
      queue:  document.getElementById('step-queue'),
      worker: document.getElementById('step-worker'),
      thumb:  document.getElementById('step-thumb'),
      done:   document.getElementById('step-done')
    };

    // Restore a previously entered key for convenience
    var savedKey = localStorage.getItem('dpp_api_key');
    if (savedKey) { apiKeyInput.value = savedKey; }

    // ── File type selector ──────────────────────────────
    document.querySelectorAll('#type-seg button').forEach(function(btn) {
      btn.addEventListener('click', function() {
        document.querySelectorAll('#type-seg button').forEach(function(b){ b.classList.remove('active'); });
        btn.classList.add('active');
        typeMode = btn.getAttribute('data-type');
        fileInput.accept = typeMode === 'md' ? '.md' : typeMode === 'docx' ? '.docx' : '.md,.docx';
        if (typeMode === 'md')        { dropMeta.textContent = '.md only · max 5 MB'; }
        else if (typeMode === 'docx') { dropMeta.textContent = '.docx only · max 20 MB'; }
        else                          { dropMeta.textContent = '.md (max 5 MB) or .docx (max 20 MB)'; }
        if (selectedFile) { setFile(selectedFile); } // re-validate against the new mode
      });
    });

    function detectType(name) {
      var n = (name || '').toLowerCase();
      if (n.endsWith('.md'))   { return 'md'; }
      if (n.endsWith('.docx')) { return 'docx'; }
      return null;
    }

    fileInput.addEventListener('change', function(e){ setFile(e.target.files[0]); });
    dropZone.addEventListener('dragover', function(e){ e.preventDefault(); dropZone.classList.add('drag-over'); });
    dropZone.addEventListener('dragleave', function(){ dropZone.classList.remove('drag-over'); });
    dropZone.addEventListener('drop', function(e){
      e.preventDefault(); dropZone.classList.remove('drag-over');
      var file = e.dataTransfer.files[0]; if (file) { setFile(file); }
    });
    chipRemove.addEventListener('click', function(e){ e.stopPropagation(); resetFile(); });

    function setFile(file) {
      if (!file) { return; }
      var detected = detectType(file.name);
      var type = typeMode === 'auto' ? detected : typeMode;
      if (!type) {
        showError('Unsupported file. Use a .md or .docx document.'); return;
      }
      if (typeMode !== 'auto' && detected && detected !== typeMode) {
        showError('You selected ' + typeMode.toUpperCase() + ' but the file looks like .' + detected + '.'); return;
      }
      var maxMB = type === 'docx' ? 20 : 5;
      if (file.size > maxMB * 1024 * 1024) {
        showError('File too large — max ' + maxMB + ' MB for ' + type.toUpperCase() + '.'); return;
      }
      selectedFile = file;
      currentType = type;
      chipName.textContent = file.name;
      chipType.textContent = type;
      fileChip.classList.add('show');
      uploadBtn.disabled = false;
      btnText.textContent = 'Convert to PDF';
      hideError();
    }

    function resetFile() {
      selectedFile = null; currentType = null; fileInput.value = '';
      fileChip.classList.remove('show'); uploadBtn.disabled = true;
      btnText.textContent = 'Select a file to convert';
    }

    uploadBtn.addEventListener('click', startUpload);

    async function startUpload() {
      apiKey = apiKeyInput.value.trim();
      if (!apiKey) { showError('Enter your API key first (sent as the X-API-Key header).'); return; }
      if (!selectedFile) { return; }

      var webhookURL = webhookInput.value.trim();
      if (webhookURL && webhookURL.toLowerCase().indexOf('https://') !== 0) {
        showError('Webhook URL must start with https://'); return;
      }

      localStorage.setItem('dpp_api_key', apiKey);
      hideError(); rateBanner.classList.remove('show');
      done = false; sseOpen = false;
      uploadBtn.disabled = true; btnText.textContent = '⏳ Uploading...';
      dropZone.style.pointerEvents = 'none'; chipRemove.style.display = 'none';
      pipeHead.classList.add('show'); pipeline.classList.add('show');
      setStep('upload', 'active');

      var formData = new FormData();
      formData.append('file', selectedFile);
      if (webhookURL) { formData.append('webhook_url', webhookURL); }

      try {
        var res = await fetch('/upload', {
          method: 'POST',
          headers: { 'X-API-Key': apiKey },
          body: formData
        });
        if (res.status === 429) { rateBanner.classList.add('show'); resetToIdle(); return; }
        var data = await res.json();
        if (!res.ok) { showError(data.error || 'Upload failed.'); resetToIdle(); return; }

        currentFileId = data.file_id;
        fileIdTag.textContent = 'job_id: ' + currentFileId;
        workerDesc.textContent = 'Picked up by ' + workerLabel(currentType);
        setStep('upload', 'done'); setStep('queue', 'done'); setStep('worker', 'active');
        btnText.textContent = '⚙ Processing...';
        connectSSE(currentFileId);
      } catch (err) {
        showError('Network error — ' + err.message); resetToIdle();
      }
    }

    function workerLabel(t) {
      return t === 'docx' ? 'worker-docx (LibreOffice)' : 'worker-md (wkhtmltopdf)';
    }

    // ── SSE with polling fallback ───────────────────────
    function connectSSE(fileId) {
      if (!window.EventSource) { startPolling(fileId); return; }
      try {
        es = new EventSource('/events/' + encodeURIComponent(fileId) + '?api_key=' + encodeURIComponent(apiKey));
      } catch (e) {
        startPolling(fileId); return;
      }

      // If SSE has not connected within 5s, fall back to polling
      sseFallbackTimer = setTimeout(function() {
        if (!sseOpen && !done) { teardownSSE(); startPolling(fileId); }
      }, 5000);

      // Backstop: even while SSE is connected, poll /status slowly so the UI
      // still resolves if a notification is ever dropped (SSE open but silent).
      backstopInterval = setInterval(function() { pollOnce(fileId); }, 20000);

      es.onopen = function() {
        sseOpen = true;
        clearTimeout(sseFallbackTimer);
        setLive('sse', 'Live');
      };
      es.addEventListener('job_update', function(ev) {
        var d = null;
        try { d = JSON.parse(ev.data); } catch (e) { return; }
        handleUpdate(d);
      });
      es.onerror = function() {
        // The server closes the stream after the single terminal event; if we are
        // already done, ignore. If we never opened, fall back to polling.
        if (done) { teardownSSE(); return; }
        if (!sseOpen) { teardownSSE(); startPolling(fileId); }
      };
    }

    function teardownSSE() {
      if (sseFallbackTimer) { clearTimeout(sseFallbackTimer); sseFallbackTimer = null; }
      if (es) { es.close(); es = null; }
    }

    // One status check, shared by the fast fallback poll and the slow backstop.
    async function pollOnce(fileId) {
      try {
        var res = await fetch('/status/' + encodeURIComponent(fileId), {
          headers: { 'X-API-Key': apiKey }
        });
        var d = await res.json();
        handleUpdate(d);
      } catch (e) { /* keep trying on the next tick */ }
    }

    function startPolling(fileId) {
      if (pollInterval) { return; }
      setLive('poll', 'Polling');
      pollInterval = setInterval(function() { pollOnce(fileId); }, 2500);
    }

    function stopUpdates() {
      teardownSSE();
      if (pollInterval) { clearInterval(pollInterval); pollInterval = null; }
      if (backstopInterval) { clearInterval(backstopInterval); backstopInterval = null; }
    }

    function handleUpdate(d) {
      if (!d || !d.status || done) { return; }
      if (d.status === 'completed') {
        done = true; stopUpdates(); hideLive();
        progressToComplete(d);
      } else if (d.status === 'failed') {
        done = true; stopUpdates(); hideLive();
        setStep('worker', 'failed'); setStep('thumb', 'failed'); setStep('done', 'failed');
        showError('Conversion failed after max retries. Check the worker logs.');
        uploadBtn.style.display = ''; uploadBtn.disabled = false; btnText.textContent = 'Try Again';
        dropZone.style.pointerEvents = ''; chipRemove.style.display = '';
      }
      // 'processing' → leave the worker step spinning
    }

    function progressToComplete(d) {
      setStep('worker', 'done');
      setStep('thumb', 'active');
      setTimeout(function() {
        setStep('thumb', 'done');
        setStep('done', 'active');
        setTimeout(function() {
          setStep('done', 'done');
          showDownload(d.download_url);
          if (d.thumbnail_url) { showThumbnail(d.thumbnail_url); }
        }, 450);
      }, 600);
    }

    function showThumbnail(url) {
      thumbImg.onload = function() {
        // next frame so the transition runs
        requestAnimationFrame(function() { thumbPreview.classList.add('show'); });
      };
      thumbImg.src = url;
    }

    function showDownload(url) {
      if (url) { dlBtn.href = url; dlBtn.style.display = ''; }
      else     { dlBtn.style.display = 'none'; }
      dlBlock.classList.add('show');
      uploadBtn.style.display = 'none';
    }

    // ── Live / Polling indicator ────────────────────────
    function setLive(kind, text) {
      liveBadge.className = 'live-badge ' + kind + ' show';
      liveText.textContent = text;
    }
    function hideLive() { liveBadge.classList.remove('show'); }

    newBtn.addEventListener('click', function() {
      stopUpdates(); currentFileId = null; done = false; sseOpen = false;
      Object.keys(steps).forEach(function(k){ steps[k].className = 'pipe-step'; });
      pipeHead.classList.remove('show'); pipeline.classList.remove('show');
      dlBlock.classList.remove('show'); thumbPreview.classList.remove('show');
      thumbImg.removeAttribute('src'); hideLive();
      uploadBtn.style.display = ''; dropZone.style.pointerEvents = '';
      chipRemove.style.display = ''; resetFile(); hideError();
      rateBanner.classList.remove('show');
    });

    function setStep(name, state) { steps[name].className = 'pipe-step ' + state; }

    function resetToIdle() {
      stopUpdates(); hideLive();
      pipeHead.classList.remove('show'); pipeline.classList.remove('show');
      Object.keys(steps).forEach(function(k){ steps[k].className = 'pipe-step'; });
      uploadBtn.disabled = false; uploadBtn.style.display = '';
      btnText.textContent = selectedFile ? 'Convert to PDF' : 'Select a file to convert';
      dropZone.style.pointerEvents = ''; chipRemove.style.display = '';
    }

    function showError(msg) { errorMsg.textContent = msg; errorBanner.classList.add('show'); }
    function hideError() { errorBanner.classList.remove('show'); }
  </script>
</body>
</html>`
