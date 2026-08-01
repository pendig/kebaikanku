const WAITLIST_COOLDOWN_SECONDS = 20;
const MAX_BODY_BYTES = 1_048_576;
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export async function onRequestPost(context) {
	const { request, env } = context;

	if (!env.WAITLIST_DB) {
		return jsonError(500, 'CONFIG_ERROR', 'Waitlist database is not configured.');
	}

	const bodySize = Number(request.headers.get('content-length') || 0);
	if (bodySize > MAX_BODY_BYTES) {
		return jsonError(413, 'PAYLOAD_TOO_LARGE', 'Payload is too large.');
	}

	let payload;
	try {
		payload = await request.json();
	} catch {
		return jsonError(400, 'INVALID_JSON', 'Payload must be valid JSON.');
	}

	const email = String(payload?.email || '').trim().toLowerCase();
	if (!EMAIL_RE.test(email)) {
		return jsonError(400, 'INVALID_EMAIL', 'Email is required and must be valid.');
	}

	if (String(payload?.website || '').trim() !== '') {
		return jsonError(400, 'BOT_DETECTED', 'Request rejected as suspicious.');
	}

	const clientIP = getClientIP(request);
	const source = String(payload?.source || 'landing-coming-soon').trim().slice(0, 255);
	const userAgent = String(request.headers.get('user-agent') || '').slice(0, 1000);
	const now = new Date().toISOString();

	const limited = await isRateLimited(env.WAITLIST_DB, clientIP);
	if (limited) {
		return jsonError(429, 'RATE_LIMITED', 'Too many waitlist requests. Please try again later.');
	}

	const id = crypto.randomUUID();
	try {
		await env.WAITLIST_DB.prepare(
			`INSERT INTO waitlists (id, email, source, ip_address, user_agent, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`
		).bind(id, email, source, clientIP, userAgent, now, now).run();
	} catch (error) {
		if (isDuplicateError(error)) {
			return jsonError(409, 'DUPLICATE_EMAIL', 'Email is already registered in the waitlist.');
		}
		console.error('Failed to store waitlist signup:', error);
		return jsonError(500, 'DB_ERROR', 'Could not process waitlist registration.');
	}

	const entry = { id, email, source, ipAddress: clientIP, userAgent, createdAt: now };
	const emailQueued = Boolean(env.EMAIL && env.WAITLIST_EMAIL_FROM);
	const adminEmailQueued = emailQueued && Boolean(env.WAITLIST_ADMIN_EMAIL);

	if (emailQueued || adminEmailQueued) {
		context.waitUntil(deliverWaitlistEmails(env, entry));
	}

	return jsonResponse(201, {
		success: true,
		data: {
			id,
			email_queued: emailQueued,
			admin_email_queued: adminEmailQueued
		}
	});
}

export function onRequestOptions() {
	return new Response(null, { status: 204 });
}

async function isRateLimited(db, clientIP) {
	if (!clientIP) return false;

	const cutoff = new Date(Date.now() - WAITLIST_COOLDOWN_SECONDS * 1000).toISOString();
	const result = await db
		.prepare('SELECT id FROM waitlists WHERE ip_address = ? AND created_at >= ? LIMIT 1')
		.bind(clientIP, cutoff)
		.first();

	return Boolean(result);
}

async function deliverWaitlistEmails(env, entry) {
	const tasks = [];

	if (env.EMAIL && env.WAITLIST_EMAIL_FROM) {
		tasks.push(sendWaitlistConfirmation(env, entry.email));
	}
	if (env.EMAIL && env.WAITLIST_EMAIL_FROM && env.WAITLIST_ADMIN_EMAIL) {
		tasks.push(sendWaitlistAdminNotification(env, entry));
	}

	const results = await Promise.allSettled(tasks);
	for (const result of results) {
		if (result.status === 'rejected') {
			console.error('Failed to deliver waitlist email:', result.reason);
		}
	}
}

function sendWaitlistConfirmation(env, recipient) {
	const subject = 'Anda sudah masuk waitlist kebaikanku.id';
	const body = `Halo,

Terima kasih sudah bergabung ke waitlist kebaikanku.id.

Kami akan mengirimkan update saat dashboard pengelola kampanye dan integrasi payment gateway sudah siap dirilis.

Pantau halaman ini untuk update berikutnya:
${env.WAITLIST_EMAIL_URL || 'https://kebaikanku.id/coming-soon'}

Salam,
Tim kebaikanku.id
`;

	return sendEmail(env, recipient, subject, body);
}

function sendWaitlistAdminNotification(env, entry) {
	const subject = 'Waitlist baru kebaikanku.id';
	const body = `Ada pendaftar baru di waitlist kebaikanku.id.

Email: ${entry.email}
Source: ${entry.source}
IP Address: ${entry.ipAddress}
User Agent: ${entry.userAgent}
Created At: ${entry.createdAt}
ID: ${entry.id}
`;

	return sendEmail(env, env.WAITLIST_ADMIN_EMAIL, subject, body);
}

function sendEmail(env, recipient, subject, body) {
	return env.EMAIL.send({
		from: env.WAITLIST_EMAIL_FROM,
		to: recipient,
		subject,
		text: body
	});
}

function getClientIP(request) {
	return request.headers.get('cf-connecting-ip') || '';
}

function isDuplicateError(error) {
	return String(error?.message || error).toLowerCase().includes('unique');
}

function jsonError(status, code, message) {
	return jsonResponse(status, {
		success: false,
		error: { code, message }
	});
}

function jsonResponse(status, data) {
	return Response.json(data, {
		status,
		headers: {
			'Cache-Control': 'no-store'
		}
	});
}
