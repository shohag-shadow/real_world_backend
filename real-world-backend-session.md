# From CRUD to a Real App: A 9 Day Mentoring Session

## Contents

| Day | Topic | Homework answers | Lesson | Build task | Research | Questions |
| - | - | - | - | - | - | - |
| [Day 1](#day-1-who-are-you-authentication-foundations) | Who Are You? Authentication Foundations | n/a | [Lesson](#todays-goal) | [Build](#build-task-for-day-1) | [Research](#research-list-for-day-2) | [Questions](#questions-for-day-2-write-your-answers-first) |
| [Day 2](#day-2-staying-logged-in-safely-and-every-kind-of-authentication) | Staying Logged In Safely, and Every Kind of Authentication | [Answers](#homework-answers-from-day-1) | [Lesson](#todays-lesson) | [Build](#build-task-for-day-2) | [Research](#research-list-for-day-3) | [Questions](#questions-for-day-3) |
| [Day 3](#day-3-what-may-you-do-authorization) | What May You Do? Authorization | [Answers](#homework-answers-from-day-2) | [Lesson](#todays-lesson-1) | [Build](#build-task-for-day-3) | [Research](#research-list-for-day-4) | [Questions](#questions-for-day-4) |
| [Day 4](#day-4-when-things-fail-halfway-database-transactions-and-concurrency) | When Things Fail Halfway: Database Transactions and Concurrency | [Answers](#homework-answers-from-day-3) | [Lesson](#todays-lesson-2) | [Build](#build-task-for-day-4) | [Research](#research-list-for-day-5) | [Questions](#questions-for-day-5) |
| [Day 5](#day-5-handling-money-without-losing-it) | Handling Money Without Losing It | [Answers](#homework-answers-from-day-4) | [Lesson](#todays-lesson-3) | [Build](#build-task-for-day-5) | [Research](#research-list-for-day-6) | [Questions](#questions-for-day-6) |
| [Day 6](#day-6-doing-work-later-queues-and-background-jobs) | Doing Work Later: Queues and Background Jobs | [Answers](#homework-answers-from-day-5) | [Lesson](#todays-lesson-4) | [Build](#build-task-for-day-6) | [Research](#research-list-for-day-7) | [Questions](#questions-for-day-7) |
| [Day 7](#day-7-telling-people-what-happened-notifications) | Telling People What Happened: Notifications | [Answers](#homework-answers-from-day-6) | [Lesson](#todays-lesson-5) | [Build](#build-task-for-day-7) | [Research](#research-list-for-day-8) | [Questions](#questions-for-day-8) |
| [Day 8](#day-8-putting-it-all-together-ticketbox-end-to-end) | Putting It All Together: TicketBox End to End | [Answers](#homework-answers-from-day-7) | [Lesson](#todays-lesson-designing-ticketbox-properly) | [Build](#final-build-task) | [Research](#research-list-final) | [Questions](#final-questions-answers-in-day-9) |
| [Day 9](#day-9-final-answers-and-what-comes-next) | Final Answers and What Comes Next | [Answers](#answers-to-the-day-8-questions) | n/a | n/a | n/a | n/a |

Before Day 1: [How this session works](#how-this-session-works), [The running project: TicketBox](#the-running-project-ticketbox), [The shape of a real request](#the-shape-of-a-real-request).

After Day 9: [Closing notes from me](#closing-notes-from-me).

Each day (from Day 2) opens with my answers to the previous day's questions, so do your own research first.


You already know how to build a CRUD API and a frontend that talks to it. That is a real skill, and it is the base everything else sits on. But a CRUD app assumes that everyone is honest, every request succeeds, every operation is fast, and nobody ever clicks a button twice. Real apps get none of those favors.

I am going to teach you this the way I would teach a junior engineer on my team. We will cover five problems that every serious application has to solve:

1. **Authentication**: who are you?
2. **Authorization**: what are you allowed to do?
3. **Transactions**: what happens when something fails halfway, especially with money?
4. **Queues**: what work should happen later, in the background?
5. **Notifications**: how do we tell people what happened?

***

## How this session works

Each day has the same shape:

* **Homework answers** (from Day 2 onward): I answer the questions I gave you the day before, and I share my own version of the research I asked you to do, plus what I consider best practice.
* **Today's lesson**: the concepts, with real-life examples, small code sketches, and common mistakes.
* **Build task**: something small to implement so the idea sticks.
* **Research list**: topics to read about on your own.
* **Questions**: things you should be able to answer after your research. I answer them at the start of the next day.

**The rule that makes this work:** do the research and write down your own answers *before* you read my answers. If you read mine first, you learn nothing about where your thinking has gaps. Being wrong in your notes is fine. Being wrong in production is expensive, so this is the cheap place to be wrong.

### The language does not matter

I will use short pseudocode and SQL. The ideas are identical in every framework. Here is where to look in the ecosystems you are likely to touch:

| Concept | Node | .NET | Laravel | Go |
| - | - | - | - | - |
| Password hashing | argon2, bcrypt packages | ASP.NET Core Identity hasher | Hash facade | x/crypto argon2 and bcrypt |
| Login and tokens | Passport, jose | JWT bearer, Identity | Sanctum, Passport | golang-jwt, middleware |
| Authorization | CASL | Policy based authorization | Gates and Policies | Casbin |
| Queues | BullMQ | Hangfire, MassTransit | Laravel Queues | Asynq, River |
| Notifications | provider SDKs | MailKit, provider SDKs | Notifications | provider SDKs |

Learn the concept first. The framework is just a convenient wrapper around it.

### The running project: TicketBox

We will design one app across all days, so each lesson has a place to land. **TicketBox** is a concert ticket booking app.

* **Customers** register, log in, browse events, buy tickets, request refunds.
* **Organizers** create events, set ticket types and prices, see their own sales.
* **Admins** manage users, handle disputes, view everything.
* Customers hold a **wallet balance** and can also pay by card through an external payment provider.
* Buying a ticket creates a **receipt email**, a **confirmation notification**, and a **reminder** the day before the event.
* Tickets are limited. When two people click "buy" for the last ticket at the same moment, exactly one must win.

Every hard topic in this session shows up in that one sentence about the last ticket.

### The shape of a real request

Keep this picture in your head. We will fill it in over nine days.

```
Browser or mobile app
        |
        v
  HTTPS, rate limiting                 (Day 2)
        |
        v
  Authentication: who is this?         (Days 1, 2)
        |
        v
  Authorization: may they do this?     (Day 3)
        |
        v
  Validation, then business logic
        |
        v
  Database transaction                 (Days 4, 5)
        |            \
        v             v
   HTTP response    Outbox / queue     (Day 6)
                          |
                          v
                   Background worker
                          |
                          v
          Email, push, SMS, in-app     (Day 7)
```

***

# Day 1: Who Are You? Authentication Foundations

## Today's goal

By tonight you should understand how a server decides that a request really comes from a specific person, and why almost every beginner implementation of this is unsafe.

## 1. Authentication versus authorization

Think of a hotel. At the front desk you show your passport and they confirm you are the person who booked room 412. That is **authentication**: proving who you are. They hand you a key card. The card opens room 412 and the gym, and nothing else. That is **authorization**: what that identity may do.

Beginners blur these two. Keep them separate in your head and in your code. Today is only about the front desk.

## 2. Passwords: what the database must never contain

The first rule of storing passwords is that **you never store the password**. You store something that lets you check a password without being able to recover it.

**Hashing** is a one way function. You feed in `hunter2` and get a long string of characters. You cannot run it backwards. At login, you hash what the user typed and compare it to the stored hash.

**Encryption** is reversible by design. Whoever has the key can recover the original. That is exactly what you do not want for passwords.

But plain hashing is not enough. If you use a fast hash like SHA-256 for passwords, an attacker who steals your database can test billions of guesses per second on a graphics card. So password hashing uses algorithms that are deliberately slow and memory hungry:

* **Argon2id**: the modern recommendation. Costs both time and memory, which makes GPU attacks expensive.
* **bcrypt**: older, still widely used and acceptable. It has a limit of 72 bytes of input, which surprises people with long passphrases.
* **scrypt** and **PBKDF2**: acceptable, with more careful tuning needed.

Two more words you will hear:

* **Salt**: a random value generated per user and mixed into the hash. It means two users with the same password get different hashes, and it kills precomputed lookup tables ("rainbow tables"). Modern libraries generate and store the salt inside the hash string for you. You do not manage it manually.
* **Pepper**: a secret value kept outside the database (for example in a secrets manager) and mixed in as well. Optional extra layer. If someone steals only the database, they are missing the pepper.

### Password policy, the modern way

Older advice said "require a symbol, a number, an uppercase letter, and change it every 90 days." Current guidance (NIST's digital identity guidelines) says almost the opposite:

* Prefer **length** over complexity rules.
* **Check new passwords against lists of breached passwords.**
* Do **not** force periodic rotation without evidence of compromise. It pushes people toward `Password1`, `Password2`.
* Allow pasting into the field so password managers work.
* Normalize Unicode so the same visible password always hashes identically.

## 3. Sessions versus tokens

After the user proves who they are once, we do not want them to send their password with every request. So we give them a **proof of login** to carry. There are two main designs.

### Server side sessions

The server creates a random session ID, stores the session record on its side (database or Redis), and gives the browser the ID in a cookie. On each request, the browser sends the cookie, and the server looks the ID up.

This is like a coat check. You hand over your coat, you get a numbered ticket. The ticket means nothing by itself. The coat room knows what it maps to.

* **Pro:** you can end a session instantly, by deleting the record.
* **Con:** every request needs a lookup, and multiple servers must share the session store.

### Tokens (JWT)

A **JSON Web Token** is a string with three parts: header, payload, signature. The payload contains claims such as the user ID and an expiry time. The server signs it with a secret. Later it can verify the signature without looking anything up.

This is like a concert wristband. Security does not call the box office to check your wristband. They look at it and know it is genuine.

Critical detail that most tutorials gloss over: **a JWT is signed, not encrypted.** Anyone holding the token can read the payload. Never put secrets or sensitive personal data in it.

* **Pro:** no lookup needed to verify, works well across many services.
* **Con:** you cannot easily take it back. A wristband that has not expired stays valid.

Neither design is "better". They trade control for scalability. Most real systems end up with a hybrid, which we build tomorrow.

## 4. Cookies, and the three flags you must know

If you use cookies for login state, set these on the cookie:

* `HttpOnly`: JavaScript cannot read it. This limits the damage of injected scripts (XSS).
* `Secure`: only sent over HTTPS.
* `SameSite=Lax` or `Strict`: the browser will not attach it to most cross site requests. This is your main protection against CSRF, where another website tricks the user's browser into sending a request to yours.

## 5. Small details that separate real implementations from tutorials

* **Normalize the email** (trim whitespace, lowercase) before storing and before lookup.
* **Same error for wrong email and wrong password.** "Invalid email or password." Never tell the caller which part was wrong.
* **Same timing, too.** If an unknown email returns in 2 ms and a known email takes 150 ms (because hashing runs), an attacker can tell which accounts exist. Run a hash against a dummy value for unknown emails.
* **Regenerate the session ID after login.** Otherwise an attacker who planted a known session ID before login (session fixation) inherits the logged in session.
* **Never log passwords**, including in error logs and request logs. Check that your request logger does not dump bodies for `/login`.
* **Use HTTPS everywhere.** Without it, nothing else matters.

## 6. Sketch: registration and login

```
function register(email, password):
    email = normalize(email)
    validate(password)                        # length, not in breached list
    hash = argon2id(password)
    try:
        insert into users (email, password_hash) values (email, hash)
    catch UniqueViolation:
        return 409                            # see question 6 below
    return 201

function login(email, password):
    user = find user by normalize(email)
    if user is missing:
        verify(password, DUMMY_HASH)          # spend the same time
        return 401 "Invalid email or password"
    if not verify(password, user.password_hash):
        return 401 "Invalid email or password"
    return create_session_or_token(user)
```

```sql
CREATE TABLE users (
  id                BIGSERIAL PRIMARY KEY,
  email             TEXT NOT NULL,
  password_hash     TEXT NOT NULL,
  email_verified_at TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

/* uniqueness enforced by the database, case insensitively */
CREATE UNIQUE INDEX users_email_unique ON users (lower(email));
```

## Common mistakes to avoid

* Using MD5, SHA-1, or plain SHA-256 for passwords.
* Writing your own cryptography. Use the vetted library.
* Putting a long lived JWT in `localStorage` and calling it done.
* Returning the whole user row, including `password_hash`, from an API.
* Confirming "this email is not registered" on the login form.

## Build task for Day 1

Build `POST /register`, `POST /login`, and `GET /me`. Hash with Argon2id or bcrypt. Pick either sessions or a short lived JWT for `/me`. Add the unique index. Then try to break your own app: register the same email in two different cases, send a wrong password and an unknown email and compare response times.

## Research list for Day 2

1. How password hashing works: salt, work factor, and why Argon2id is memory hard. Read the OWASP Password Storage Cheat Sheet.
2. Sessions versus JWT: how real products decide between them.
3. What XSS and CSRF are, and how cookie flags relate to each.
4. HTTP status codes 401 and 403.
5. Credential stuffing: what it is and why reused passwords make it work.

## Questions for Day 2 (write your answers first)

1. Why can't we simply encrypt passwords instead of hashing them?
2. A JWT is stateless. So how do you log a user out, or ban them, immediately?
3. Where should a browser app keep its token: `localStorage` or an `HttpOnly` cookie? What are the trade-offs?
4. Why should a login form never say "no account with this email"?
5. What is the difference between a 401 and a 403? Which do you return when someone is not logged in?
6. Two people try to register the same email at the exact same moment. How do you guarantee only one account is created? (An `if` check in code is not the answer.)

***

# Day 2: Staying Logged In Safely, and Every Kind of Authentication

## Homework answers from Day 1

### Q1. Why not encrypt passwords instead of hashing them?

Encryption can be undone with a key. Keys get stolen, leaked in logs, or are simply known by too many employees. If passwords are encrypted, one stolen key exposes every user's real password, and people reuse passwords across sites, so that breach spreads beyond your app. With hashing there is no key to steal. Even your own administrators cannot read a user's password, which is the property we want.

**My research notes:** the fast versus slow distinction matters most. A fast hash lets an attacker test enormous numbers of guesses on a GPU. Argon2id wins the Password Hashing Competition and forces each guess to use real memory, which makes mass guessing costly. **Best practice:** use Argon2id, or bcrypt if your platform lacks it, through a well maintained library. Follow the current OWASP parameters, and store the algorithm settings inside the hash string so you can **rehash on login** when you raise the cost later.

### Q2. How do you log out or ban a user with stateless JWTs?

You cannot fully, and that is the honest answer. Stateless means the server does not remember tokens, so it cannot forget them either. The practical options are:

* **Short lived access tokens** (5 to 15 minutes) paired with a **refresh token** that is stored server side. Logging out or banning means deleting the refresh token. The worst case is a few minutes of leftover access.
* A **denylist** of token IDs (`jti` claim) kept in Redis until each token's natural expiry. Works, but now you do a lookup on every request, which partly defeats the point of statelessness.
* A **token version** on the user row. Bump it to invalidate all old tokens. Also requires a lookup.

**Best practice:** short access token plus revocable refresh token, which is what we build today. For high risk actions (changing a password, moving money), check the database anyway.

### Q3. `localStorage` or `HttpOnly` cookie?

* `localStorage` is readable by any JavaScript on the page. One XSS bug, or one compromised third party script, and the attacker copies the token.
* An `HttpOnly` cookie cannot be read by JavaScript, so XSS cannot steal it. But browsers attach cookies automatically, which opens the door to CSRF. You counter that with `SameSite` and, where needed, CSRF tokens.

Even with `HttpOnly`, XSS can still make requests *as the user* while the page is open. It just cannot carry the token away to use later. So you still must prevent XSS.

**Best practice for browser apps:** keep the refresh token in an `HttpOnly; Secure; SameSite` cookie, keep the short lived access token in memory only. For native mobile apps, use the platform's secure storage (Keychain on iOS, Keystore on Android).

### Q4. Why never say "no account with this email"?

Because it lets anyone test whether a person has an account on your service. That is a privacy leak on its own, and it feeds targeted phishing and credential stuffing. The fix is one generic message and equal response time.

**Subtle part:** the same leak hides in other places. Password reset ("we sent you an email" even if no account exists), and registration ("email already taken"). For reset, always respond the same way and send mail only if the account exists. For registration, you cannot fully hide it without an email first flow, so accept some leakage there but rate limit it.

### Q5. 401 versus 403

* **401 Unauthorized** really means *unauthenticated*: I do not know who you are, or your credentials are missing or invalid. The name is a historical mistake.
* **403 Forbidden** means I know exactly who you are, and you are not allowed.

Not logged in: 401. Logged in but wrong role: 403. A third case on Day 3: sometimes you return **404** on purpose so you do not reveal that a resource exists.

### Q6. Two identical registrations at the same instant

Check then insert is a **race condition**: both requests check, both see "no such email," both insert. The only reliable fix is a **unique index in the database**. The database serializes the writes and rejects the second one, and your code catches the unique violation and returns 409. Rule to remember for life: *any invariant that matters must be enforced by the database, not only by application code.* Day 4 is about this idea.

***

## Today's lesson

### 1. The access token plus refresh token pattern

Here is a real life picture. A day pass at a theme park expires at closing. You also hold a season membership card that you keep safely in your wallet. When the day pass expires, you show the membership card at the gate to get a fresh day pass. If someone steals your day pass, it dies tonight. The membership card is the one you guard carefully.

```
Login:
  client -> server : email + password
  server -> client : access token (15 min) + refresh token (30 days)

Normal request:
  client -> API : Authorization: Bearer <access token>

When the access token expires:
  client -> /auth/refresh : refresh token
  server -> client : NEW access token + NEW refresh token (the old refresh token is now dead)

Logout:
  server deletes the refresh token record
```

### 2. Refresh token rotation and reuse detection

Each time a refresh token is used, issue a new one and invalidate the old one. If an *already used* refresh token shows up again, that means two parties hold the same token, one of them an attacker. Treat it as theft: revoke the whole chain of tokens for that login and force a fresh login.

Store refresh tokens in a table:

```sql
CREATE TABLE refresh_tokens (
  id            BIGSERIAL PRIMARY KEY,
  user_id       BIGINT NOT NULL REFERENCES users(id),
  token_hash    TEXT NOT NULL UNIQUE,     /* store a hash, never the raw token */
  family_id     UUID NOT NULL,            /* all rotations of one login share this */
  expires_at    TIMESTAMPTZ NOT NULL,
  used_at       TIMESTAMPTZ,              /* set when rotated; reuse after this = theft */
  revoked_at    TIMESTAMPTZ,
  user_agent    TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

The raw token is 32 random bytes from the operating system's secure random generator. Because it is truly random and long, a fast hash such as SHA-256 is fine for storing it. Slow hashing is only needed for human chosen secrets.

### 3. JWT details that matter

A well formed access token carries these claims, and your verifier must check each:

* `sub`: who it is about (user id)
* `exp`: expiry. Always set, always check
* `iat`: issued at
* `iss`: who issued it. Check it matches you
* `aud`: who it is for. Check it matches your API
* `jti`: unique token id, useful for denylists
* custom claims: keep few and small

Verification mistakes that have caused real breaches:

* **Trusting the `alg` header.** An attacker sets `alg: none` or switches from RS256 to HS256 and signs with the public key. Always configure the verifier with the **one algorithm you expect**.
* **Not checking `exp`, `iss`, or `aud`.**
* **Putting authorization data in the token and trusting it for too long.**

HS256 uses one shared secret for signing and verifying. RS256 or ES256 use a private key to sign and a public key to verify, which lets other services verify tokens without being able to forge them. Choose asymmetric when more than one party verifies.

### 4. Email verification and password reset

These flows look simple and are attack magnets. The rules:

* The link carries a **random single use token**, 32 bytes or more, from a secure random generator.
* Store only its **hash** in the database, plus an expiry (15 to 60 minutes for reset).
* **Single use.** Mark it consumed in the same transaction that changes the password.
* Respond identically whether or not the account exists.
* After a password change, **revoke all refresh tokens** and email the user that it happened.
* Never put the reset token in a URL that gets logged by third party scripts. Keep reset pages free of analytics scripts and set a `Referrer-Policy` header.

### 5. Brute force protection

* Rate limit login by **IP address** and by **account**, separately. IP alone fails against botnets; account alone lets an attacker lock out real users.
* Slow down with growing delays instead of hard lockouts, since hard lockouts become a denial of service tool.
* Add CAPTCHA only after suspicious behavior, not for everyone.
* Alert on a burst of failed logins across many accounts: that is credential stuffing.

### 6. A map of authentication types

You said you want the full picture, so here it is. Each type solves a different problem.

| Type | Typical use | The detail people miss |
| - | - | - |
| Password plus session cookie | Traditional web apps | Regenerate session ID on login; set cookie flags |
| Short JWT plus refresh token | SPAs, mobile apps, APIs | Rotation and reuse detection; algorithm pinning |
| Opaque access token | APIs needing instant revocation | Server looks it up every time (introspection) |
| API key | Server to server, developer access | Identifies an app, not a person. Store hashed, show once, add a visible prefix, allow rotation, scope narrowly |
| OAuth 2.0 | Letting an app act on your behalf | It is about delegated *access*, not login by itself |
| OpenID Connect (OIDC) | "Sign in with Google" style login | Adds an ID token on top of OAuth. Use the authorization code flow with PKCE |
| SAML | Enterprise single sign on | Older, XML based, still common with big companies |
| TOTP (authenticator app) | Second factor | Codes derive from a shared secret and the current time; needs recovery codes |
| WebAuthn and passkeys | Phishing resistant login | The credential is bound to your domain, so a fake site cannot use it |
| SMS one time code | Second factor of last resort | Weakest option: SIM swapping and interception exist |
| Magic link or email code | Passwordless login | Email security scanners may "click" links and burn one time tokens. Make the link open a page with a confirm button that submits a POST |
| Mutual TLS | Service to service | Both sides present certificates |

**Build versus buy.** For login with Google, MFA, and enterprise SSO, use a proven library or a managed identity provider instead of coding the protocol yourself. Understanding how it works is for debugging and decision making. Writing your own OAuth server is rarely wise.

### 7. OAuth 2.0 and OIDC in plain English

Imagine a valet key for a car. It starts the engine and opens the door but not the trunk or glovebox. OAuth is a way to hand a limited key to an app without giving it your real password. "Sign in with Google" feels like login, but OAuth itself only says "this app may access these things." **OpenID Connect** adds a signed ID token that says who the user is. If you use plain OAuth to log people in, you are using a tool for something it was not designed for.

The safe flow for web and mobile apps is **authorization code with PKCE**. The app sends the user to the provider, the provider sends back a short code, and the app exchanges that code (plus a secret it generated called the code verifier) for tokens. PKCE ensures a stolen code is useless without the verifier.

## Best practices checklist for authentication

* Argon2id or bcrypt. Breached password check. No forced rotation.
* Generic errors, equal timing, rate limits.
* Short access tokens, rotating refresh tokens, hashed storage.
* Pin the JWT algorithm and validate `exp`, `iss`, `aud`.
* `HttpOnly; Secure; SameSite` cookies.
* MFA available. Recovery codes issued.
* Password change revokes sessions and emails the user.
* Security events logged (login, failed login, reset, MFA change) without logging secrets.

## Build task for Day 2

Extend your Day 1 app with: access token (15 min), refresh token with rotation and reuse detection, `POST /logout`, password reset by email (print the link to the console for now), and a rate limit on `/login`. Write a test that reuses an old refresh token and asserts the whole family is revoked.

## Research list for Day 3

1. OAuth 2.0 authorization code flow with PKCE, and what an OIDC ID token contains.
2. How TOTP codes are generated (HMAC and time steps).
3. HS256 versus RS256, and the `alg: none` attack.
4. CORS: what it is, what it protects, and what it does not.
5. The OWASP Authentication and Session Management cheat sheets.

## Questions for Day 3

1. Why do we store only a hash of refresh tokens and reset tokens?
2. What is refresh token rotation, and what attack does reuse detection catch?
3. When does the HS256 versus RS256 choice actually matter?
4. Is OAuth 2.0 an authentication protocol? Explain.
5. A user changes their password. What should happen to their existing sessions and tokens?
6. Does CORS protect your API from attackers?

***

# Day 3: What May You Do? Authorization

## Homework answers from Day 2

### Q1. Why store only a hash of refresh tokens and reset tokens?

If your database leaks, an attacker who finds raw tokens can use them immediately. If you stored only hashes, the leaked rows are useless, because the attacker cannot turn a hash back into the token. This is the same logic as passwords, with one difference: these tokens are long random values, so guessing is already impossible and a fast hash like SHA-256 is enough. Slow hashing exists to protect weak, human chosen secrets.

### Q2. Rotation and reuse detection

Rotation means every use of a refresh token returns a new one and kills the old one. Reuse detection is the alarm: if a dead token is presented again, either the real user or a thief is replaying it. The server cannot know which, so it revokes the entire token family and forces re-login. It catches the case where an attacker steals a refresh token and the legitimate client later (or earlier) uses the same one. Without rotation, a stolen refresh token works silently for its whole 30 day life.

**My research notes:** this is the approach recommended in the OAuth 2.0 security best practices for public clients. One edge case to design for: two browser tabs refreshing at the same moment can look like reuse. Good implementations allow a very short grace window (a few seconds) for the immediately previous token, or the client serializes refresh calls.

### Q3. HS256 versus RS256

HS256 uses one shared secret. Anyone who can *verify* tokens can also *create* them. That is fine when one service both issues and verifies. It becomes dangerous the moment a second service, a partner, or a mobile app needs to verify, because you must hand them a secret that can forge tokens. RS256 and ES256 split the roles: only the auth service holds the private key, everyone else gets the public key (usually published at a JWKS URL, which also makes key rotation easy through the `kid` header). **Best practice:** HS256 for a single small service, asymmetric for anything with more than one verifier. In both cases, pin the algorithm in the verifier.

### Q4. Is OAuth 2.0 an authentication protocol?

No. OAuth 2.0 is an authorization framework: it lets a user grant an app limited access to something. An access token proves "this app may do X," not "this user is Y." Logging in with plain OAuth leads to real bugs, such as accepting a token that was issued to a different app. OpenID Connect is the identity layer built on top, giving you a signed ID token with the user's identity and checks for audience and nonce. If you want "Sign in with Google," you are using OIDC.

### Q5. User changes their password

Revoke all refresh tokens for that user (optionally keeping the one used for the current device), invalidate any outstanding reset tokens, and email the user "your password was changed" with a link to report it. If you use access tokens that live 15 minutes, they will expire on their own, and for very sensitive apps you can bump a token version to cut them off at once. Also require re-authentication (password or MFA) before allowing the change.

### Q6. Does CORS protect your API?

No. CORS is a browser rule that decides whether JavaScript from one website may *read* responses from another. A script, `curl`, or a malicious server does not care about CORS at all. Your API must authenticate and authorize every request regardless. Two practical points: never combine `Access-Control-Allow-Origin: *` with credentialed requests, and keep an explicit allowlist of origins. CORS is also separate from CSRF, which is handled by `SameSite` cookies and CSRF tokens.

***

## Today's lesson

### 1. Four layers of authorization

Picture an office building:

1. **Entrance level:** you need a badge to get into the building at all. (Route level: "must be logged in" or "must be an admin".)
2. **Floor level:** your badge opens floors 3 and 4 only. (Action level: "can this person approve refunds?")
3. **Desk level:** you may open *your* desk drawer, not your neighbor's. (Object level: "is this order mine?")
4. **Document level:** even in your own folder, some pages are blacked out. (Field level: "may this user see the cost price?")

Beginners implement layer 1 and stop. Most real security bugs live in layer 3.

### 2. The number one API bug: Broken Object Level Authorization

Say a customer calls `GET /orders/1042` and sees their order. They change the number to `1043`. If the server only checks "is logged in," they now read a stranger's order. This is often called **IDOR** (insecure direct object reference) or **BOLA**, and it tops the OWASP API Security list year after year.

The fix is never "use random IDs." Random IDs reduce guessing but do not authorize anything. The fix is to **check ownership on every object access**:

```sql
/* the safe shape: ownership is part of the query itself */
SELECT * FROM orders WHERE id = :id AND user_id = :current_user_id;
/* no row means 404, whether it is missing or belongs to someone else */
```

### 3. Role Based Access Control (RBAC)

RBAC groups permissions into roles and assigns roles to users. In TicketBox: `customer`, `organizer`, `admin`. The schema is a few small tables:

```sql
CREATE TABLE roles        (id SERIAL PRIMARY KEY, name TEXT UNIQUE NOT NULL);
CREATE TABLE permissions  (id SERIAL PRIMARY KEY, name TEXT UNIQUE NOT NULL);  /* 'event.create', 'order.refund' */
CREATE TABLE role_permissions (role_id INT REFERENCES roles(id), permission_id INT REFERENCES permissions(id),
                               PRIMARY KEY (role_id, permission_id));
CREATE TABLE user_roles   (user_id BIGINT REFERENCES users(id), role_id INT REFERENCES roles(id),
                           PRIMARY KEY (user_id, role_id));
```

**The most important RBAC habit:** your code checks **permissions**, not role names.

```
BAD :  if user.role == "admin" or user.role == "support": allow refund
GOOD:  if can(user, "order.refund", order): allow refund
```

Roles change as the business grows ("we need a support lead role"). Permissions are the stable vocabulary of what the system can do. When code checks permissions, a new role is a data change, not a code change across 40 files.

### 4. Policies: one home for each rule

Put the rules in one place per resource, not scattered across controllers.

```
policy Order:
    view(user, order):    order.user_id == user.id
                          or can(user, "order.view_any")
                          or (can(user, "order.view_own_events") and order.event.organizer_id == user.id)
    refund(user, order):  can(user, "order.refund")
                          and order.status == "paid"
                          and order.event.starts_at > now()
```

Controllers then do one thing: `authorize(user, "refund", order)`. Laravel calls these Policies and Gates, .NET calls them policy based authorization, Casbin works in several languages. Same idea.

### 5. When RBAC runs out: ABAC and relationship based access

* **ABAC** (attribute based) decides using attributes of the user, the resource, and the context. Example: "a manager can approve expenses under 500 for their own department, during working hours."
* **ReBAC** (relationship based) decides using relationships, the way file sharing works. "You can view this document if you are its owner, or an editor of it, or a member of a group it was shared with." Google published a paper on its internal system for this called **Zanzibar**, and open source systems inspired by it (OpenFGA, SpiceDB) exist if you ever need that scale.

For TicketBox, RBAC plus ownership checks is enough. Know the other two exist.

### 6. Multi-tenancy and data scoping

An organizer must only see their own events and sales. The cleanest approach is to make scoping automatic: every query for organizer data includes `organizer_id = current_user`. Some teams push this into the database using **row level security** in PostgreSQL, so even a forgotten `WHERE` clause cannot leak rows. Defense in depth means that one mistake in code does not become a breach.

### 7. Mass assignment: the quiet privilege escalation

If your code does `user.update(request.body)`, an attacker sends `{"name": "Sam", "role": "admin"}` and promotes themselves. Never bind raw request bodies into models. Use an **allowlist** of fields per endpoint (DTOs, form requests, `fillable` fields).

### 8. Where to enforce, and where roles live

* **Always enforce on the server.** The frontend hiding a "Delete" button is user experience, not security. Anyone can call the API directly.
* **Deny by default.** A new endpoint with no rule should be closed, not open.
* **Least privilege.** Give the smallest set of permissions that works.
* **Roles in the token or in the database?** Roles inside a JWT are fast to read but stale until the token expires. If an admin demotes someone, the old token still says "admin" for up to 15 minutes. For ordinary actions many teams accept that. For dangerous actions, load permissions from the database (with a short cache) at request time.
* **Audit sensitive actions.** Who refunded this order, when, and from which IP? Write it to an append only audit table. When something goes wrong, this table is your memory.

### 9. Testing authorization

Authorization bugs hide because the happy path works. Write a matrix test: for each endpoint, try it with no token, with a customer, with an organizer who owns the resource, with an organizer who does not, and with an admin. Assert the exact status code for each. This single table of tests catches regressions that code review misses.

| Endpoint | No token | Customer | Owner organizer | Other organizer | Admin |
| - | - | - | - | - | - |
| GET /events/:id | 200 | 200 | 200 | 200 | 200 |
| PATCH /events/:id | 401 | 403 | 200 | 404 or 403 | 200 |
| GET /orders/:id | 401 | own: 200, other: 404 | event owner: 200 | 404 | 200 |
| POST /orders/:id/refund | 401 | 403 | 200 | 404 or 403 | 200 |

## Best practices checklist for authorization

* Check permissions in code, map roles to permissions in data.
* Check ownership on every object access, inside the query.
* Allowlist writable fields.
* Deny by default. Enforce on the server. Frontend checks are cosmetic.
* Audit log for sensitive actions.
* Matrix tests for every protected endpoint.

## Build task for Day 3

Add roles and permissions to TicketBox. Implement the Order and Event policies above. Write the matrix test. Then deliberately attack your own API: log in as customer A and try to read, edit, and refund customer B's order. Try sending `"role": "admin"` in a profile update.

## Research list for Day 4

1. OWASP API Security Top 10: read the first five entries.
2. Mass assignment vulnerabilities and real incidents.
3. PostgreSQL row level security.
4. Google Zanzibar paper summary, and what OpenFGA does.
5. How your framework of choice implements policies (Gates, policy handlers, Casbin).

## Questions for Day 4

1. Why should code check permissions instead of role names?
2. A user requests `GET /orders/123`, which belongs to someone else. Do you return 403 or 404? Why?
3. Roles live inside the JWT. An admin demotes a user. What happens, and how would you handle it?
4. What is mass assignment, and how does it turn into privilege escalation?
5. Give an example of a rule RBAC cannot express cleanly.
6. How would you test that authorization is never accidentally bypassed?

***

# Day 4: When Things Fail Halfway: Database Transactions and Concurrency

## Homework answers from Day 3

### Q1. Permissions instead of role names

Role names describe *who someone is*, and organizations keep reshaping that: "support lead," "regional admin," "auditor." If code says `role == "admin"`, every new role means hunting through the codebase. Permissions describe *what the system can do*, and those change slowly. Code asks "may this user refund orders?" and a table decides which roles carry that permission. It also makes reviews easier: you can list exactly who can do the dangerous things.

### Q2. 403 or 404 for someone else's order?

My default is **404**. A 403 confirms "order 123 exists, it just isn't yours," which leaks information (how many orders there are, which ids are real). Scoping the query by owner returns "not found" naturally and gives the attacker nothing. There are cases where 403 is better, such as a shared workspace where users know the resource exists and need a clear message. The real rule is **be consistent and avoid leaking existence for sensitive resources.**

### Q3. Roles in the JWT after a demotion

The old token keeps claiming admin until it expires. Options, from cheapest to strongest: keep access tokens short (15 minutes); bump a token version or add the user to a short denylist; or, for dangerous actions, ignore the token's role and read permissions from the database (with a cache of a few seconds). **My practice:** short tokens for routine reads, a fresh database check for anything that moves money, deletes data, or changes permissions.

### Q4. Mass assignment

It happens when a framework copies all request fields into a model automatically. The attacker adds fields the form never showed: `role`, `is_verified`, `balance`. The fix is an explicit allowlist per endpoint. Think of it as a hotel check-in form: guests may fill in name and phone, but "room upgrade" is a field only the desk can write.

### Q5. A rule RBAC cannot express

"An organizer may refund an order only if it belongs to *their own* event and was bought *less than 7 days ago*, and only up to the amount actually paid." That mixes ownership, time, and data values: attributes, not roles. Another: "anyone who has been invited to this private event may view it." That is relationship based. In practice you handle these in policy code on top of RBAC, which is perfectly fine until the rules get large.

### Q6. Testing that authorization is never bypassed

Three layers. First, the matrix test from today for every endpoint. Second, a test that walks the router and fails if any route lacks an explicit authorization declaration (even `public`), so forgetting is impossible. Third, periodic review of the audit log and automated scans (dependency and API security scanners). The structural defense, deny by default, matters more than any single test.

***

## Today's lesson

### 1. The story that explains transactions

You send 100 to a friend. The code does two steps: subtract 100 from your balance, add 100 to theirs. The server crashes between the steps. Your money is gone and theirs never arrived. A **transaction** is the database's promise that a group of changes either all happen or none do.

### 2. ACID, with tickets

* **Atomicity:** all or nothing. Reserving a seat, creating the order, and charging the wallet succeed together or not at all.
* **Consistency:** the database never ends in a state that breaks its rules (balance must not be negative, a seat belongs to one order). Constraints enforce this.
* **Isolation:** concurrent transactions do not trample each other. Two buyers of the last ticket must behave as if they went one at a time.
* **Durability:** once the database says "committed," the data survives a crash.

### 3. Using a transaction

```sql
BEGIN;
  UPDATE wallets SET balance = balance - 5000 WHERE user_id = 1;
  INSERT INTO orders (user_id, event_id, amount) VALUES (1, 9, 5000);
COMMIT;      /* or ROLLBACK; to undo everything since BEGIN */
```

In application code, wrap this in your framework's transaction helper so that an exception causes an automatic rollback. Rules for good transaction code:

* Put the transaction boundary in the **service layer**, around one business operation, not scattered across controllers.
* Keep transactions **short**. They hold locks and a database connection.
* **No slow external calls inside** (payment provider, email, HTTP). We will see why in the homework.
* Use **constraints** as a safety net: `UNIQUE`, `CHECK (balance >= 0)`, `NOT NULL`, foreign keys. Code has bugs. Constraints do not forget.

### 4. The race condition

Seat inventory: `remaining = 1`. Two buyers, Ana and Ben, click at once. Here is the naive code and what the database sees:

```
Ana: read remaining      -> 1
Ben: read remaining      -> 1
Ana: remaining > 0, so write remaining = 0, create order
Ben: remaining > 0, so write remaining = 0, create order
Result: two tickets sold, one seat existed.
```

Notice that putting this code inside a transaction does **not** fix it at the default isolation level in most databases. Both transactions read before either writes. This is called a **lost update**.

### 5. Isolation levels in plain terms

Isolation levels are a dial between safety and speed. They are defined by which strange behaviors (anomalies) they allow:

| Level | Prevents | Still allows |
| - | - | - |
| Read Committed | Reading uncommitted data (dirty reads) | Same query returning different results within one transaction; lost updates |
| Repeatable Read | Non-repeatable reads (a consistent snapshot) | In some databases, phantoms; **write skew** |
| Serializable | All of the above: behaves as if transactions ran one by one | Costs more; transactions may be aborted and must be retried |

PostgreSQL defaults to Read Committed. MySQL's InnoDB defaults to Repeatable Read. Do not guess what you have. Check it.

**Write skew** is the subtle one. Two doctors are on call and a rule says at least one must remain. Each checks "is someone else on call?" (yes), and each marks themselves off duty. Both succeed because they updated different rows. Nobody is on call. Snapshot style isolation permits this. Serializable prevents it.

### 6. Four ways to make the last ticket safe

**Option A: a conditional atomic update (my first choice)**

```sql
UPDATE ticket_types
SET remaining = remaining - 1
WHERE id = :id AND remaining > 0;
/* if rows affected = 0, sold out. The database checks and decrements in one indivisible step. */
```

**Option B: pessimistic locking.** Lock the row first, then read and decide.

```sql
BEGIN;
SELECT remaining FROM ticket_types WHERE id = :id FOR UPDATE;   /* others wait here */
/* check, then update */
COMMIT;
```

**Option C: optimistic locking.** Add a `version` column. Update only if the version is unchanged, and retry if it was not.

```sql
UPDATE ticket_types
SET remaining = remaining - 1, version = version + 1
WHERE id = :id AND version = :version_i_read;
/* rows affected = 0 means someone else changed it; re-read and retry */
```

**Option D: Serializable isolation** and retry the whole transaction when the database aborts it.

When to use which: conditional atomic updates are simplest when the rule fits in one statement. Pessimistic locks suit busy rows where retrying is costly (wallet balances, hot inventory). Optimistic locking suits rarely contested data with long think time (two admins editing the same event description). Serializable suits complex invariants that are hard to express otherwise.

### 7. Deadlocks

Transaction 1 locks row A then wants B. Transaction 2 locks B then wants A. They wait on each other forever. The database detects this and kills one of them. Defenses: lock rows in a **consistent order** (always lower id first), keep transactions short, and **retry** a transaction that was chosen as the victim. Deadlock and serialization failures are normal events in a busy system, not mysteries.

### 8. A transaction cannot cross systems

A database transaction covers one database. It cannot include "charge this card at the payment provider" or "send this email." Anything outside the database is outside the rollback. Tomorrow and the day after, we learn the patterns for that: ledgers, idempotency keys, outbox, and compensation.

## Best practices checklist for transactions

* One business operation, one transaction, in the service layer.
* Keep them short. No network calls inside.
* Constraints for every invariant that matters.
* Conditional updates or locks for shared counters and balances.
* Retry on deadlocks and serialization failures.
* Know your database's default isolation level.
* Test concurrency with real parallel requests against a real database.

## Build task for Day 4

Implement `POST /events/:id/tickets/buy` with Option A. Seed an event with 10 tickets. Write a test that fires 50 parallel purchase requests and asserts that exactly 10 succeed, 40 fail cleanly, and `remaining` is 0 and never negative. Then deliberately write the naive version, run the same test, and watch it fail. That failure is the best lesson in this session.

## Research list for Day 5

1. What MVCC is and how PostgreSQL implements it.
2. Isolation anomalies: dirty read, non-repeatable read, phantom read, write skew.
3. `SELECT ... FOR UPDATE`, `SKIP LOCKED`, and `NOWAIT`.
4. Optimistic versus pessimistic locking.
5. How databases detect deadlocks.
6. Savepoints and nested transactions.

## Questions for Day 5

1. Why is "read remaining, check in code, then write" unsafe even inside a transaction?
2. When would you choose optimistic locking over pessimistic locking?
3. Why must you never call an external API (payments, email) inside an open database transaction?
4. What is write skew? Give an example that Repeatable Read can allow.
5. How would you prove that your purchase code is safe under concurrency?
6. What does `SKIP LOCKED` do, and where is it useful?

***

# Day 5: Handling Money Without Losing It

## Homework answers from Day 4

### Q1. Why is "read, check in code, write" unsafe inside a transaction?

Because a transaction at Read Committed only guarantees you do not see *uncommitted* data. It does not stop another transaction from reading the same value at the same time. Both buyers read `remaining = 1`, both decide "ok," both write. If your write is `SET remaining = 0` you overwrite each other (a lost update). If it is `SET remaining = remaining - 1` without a condition, you end up at -1. The check and the change must happen in **one indivisible step** inside the database: a conditional update, a row lock taken *before* the read, or a version check.

### Q2. Optimistic over pessimistic locking, when?

Optimistic locking bets that conflicts are rare. Nobody holds a lock, and a conflict is detected at write time through a version number. It fits long "think time" workflows: an organizer opens the event edit form, goes for coffee, and saves twenty minutes later. Holding a database lock for twenty minutes would be absurd. Pessimistic locking bets that conflicts are common and retries are expensive. It fits hot rows like the last tickets or a wallet balance. **My rule of thumb:** try a conditional atomic update first. If the rule does not fit in one statement, choose pessimistic for hot short operations and optimistic for slow human edits.

### Q3. Why no external calls inside a transaction?

Four reasons. (1) The transaction holds row locks and a connection for as long as the call takes, so a slow payment provider blocks every other buyer of that ticket. (2) Your connection pool drains and the whole app stalls. (3) The external side effect cannot be rolled back: if you send the email and then the commit fails, you have told the user something false. (4) If the external call times out, you do not know whether it succeeded, and you are stuck holding locks while you wonder. **Best practice:** commit your own state first, do external work after the commit or through a queue, and design for the case where it fails.

### Q4. Write skew

Two rows, one rule spanning both, and each transaction changes a different row. Doctors on call is the classic example from the last lesson. A TicketBox version: a rule "an organizer may have at most 5 active events." Two requests each count 4 active events, each create a fifth, and now the organizer has 6. A snapshot isolation level (Repeatable Read in PostgreSQL) allows this because the transactions never wrote the *same* row. Fix with Serializable isolation, by locking a parent row that represents the rule (for example, lock the organizer row), or by a constraint.

### Q5. Proving concurrency safety

You cannot prove it by reading code. You test it: spin up a real database (not a mock, since mocks do not model locks), fire many parallel requests using a load tool or test threads, and **assert the invariants**: tickets sold never exceed capacity, no balance is negative, sum of ledger entries is zero. Repeat the run many times, because races appear randomly. Add the invariant as a database constraint too, so a bug fails loudly instead of corrupting data.

### Q6. SKIP LOCKED

`SELECT ... FOR UPDATE SKIP LOCKED` locks the rows it returns and silently skips rows that other transactions have locked, instead of waiting. That lets many workers each grab a *different* unclaimed row at once. It is the core trick behind queues built on a relational database, such as the Postgres based job libraries (pg-boss is one). It is a good reminder that a database can be a perfectly fine queue at small and medium scale.

***

## Today's lesson

### 1. Rule one: never use floating point for money

In most languages, `0.1 + 0.2` is `0.30000000000000004`. Tiny errors multiply across millions of transactions. Two safe choices:

* Store **integer minor units**: 1999 for 19.99 in a currency with two decimals. Add a `currency` column (ISO 4217 code like `USD`, `EUR`, `JPY`). Remember that currencies differ: some have zero decimals, some three.
* Or use the database `NUMERIC` / `DECIMAL` type, and a decimal library in your code. Never convert through floats.

Also decide your **rounding rule** once, in one function, and write it down. Percentage fees are where rounding arguments start.

### 2. Rule two: record what happened, do not just store a balance

A column `users.balance = 4200` tells you the present and hides the past. If it is wrong, you cannot tell why. A bank does not keep only a number. It keeps a **ledger**: an append only list of every movement. The balance is a *result* of the ledger.

A bank statement is the picture to keep in mind: opening balance, then each deposit and withdrawal with a date, then a closing balance. You can always recompute the closing balance from the lines.

### 3. Double entry in plain terms

Accountants record every movement as **two sides that cancel out**. Money never appears or vanishes. It moves from one account to another. In TicketBox:

* A customer tops up 50 by card: money arrives from the *payment provider clearing account* into the *customer wallet account*.
* The customer buys a 30 ticket: 27 moves from *customer wallet* to *organizer payable*, and 3 to *platform fee revenue*.

Every transaction's entries sum to exactly zero. That one rule catches a huge class of bugs.

```sql
CREATE TABLE accounts (
  id        BIGSERIAL PRIMARY KEY,
  owner_type TEXT NOT NULL,        /* 'user', 'organizer', 'platform' */
  owner_id   BIGINT,
  kind       TEXT NOT NULL,        /* 'wallet', 'payable', 'fee_revenue', 'gateway_clearing' */
  currency   CHAR(3) NOT NULL
);

CREATE TABLE ledger_transactions (
  id            BIGSERIAL PRIMARY KEY,
  kind          TEXT NOT NULL,         /* 'topup', 'ticket_purchase', 'refund' */
  reference_type TEXT, reference_id BIGINT,  /* the order or payment this belongs to */
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ledger_entries (
  id             BIGSERIAL PRIMARY KEY,
  transaction_id BIGINT NOT NULL REFERENCES ledger_transactions(id),
  account_id     BIGINT NOT NULL REFERENCES accounts(id),
  amount_minor   BIGINT NOT NULL      /* positive = money in, negative = money out */
);
/* application rule, also verified by a test and a nightly job:
   for each transaction_id, SUM(amount_minor) = 0 */
```

Two more principles:

* **Immutable.** Never update or delete a ledger entry. To fix a mistake, post a new *reversing* entry. The history stays honest, and auditors are happy.
* **Balance from the ledger.** You may keep a cached `balance` column for fast reads, but update it in the same transaction as the entry, never independently, and reconcile it against the ledger regularly. Add `CHECK (balance >= 0)` for wallets.

### 4. Idempotency: the most important word of the week

The user taps "Pay." The network is slow. They tap again. Or your mobile client retries after a timeout. Without protection, they are charged twice.

An operation is **idempotent** if doing it twice has the same effect as doing it once. Reading is naturally idempotent. "Create a payment" is not, so we make it so with an **idempotency key**:

1. The client generates a unique key (a UUID) for each *logical* action and sends it in a header: `Idempotency-Key: 7f3c...`.
2. The server first inserts the key into a table with a **unique constraint**, scoped to the user. The database guarantees only one request wins that insert, even if two arrive simultaneously.
3. The winner does the work and stores the response against the key.
4. A repeat with the same key and same body returns the **stored response** without doing anything again. The same key with a *different* body is an error (422). A repeat while the first is still running gets a 409 "in progress".
5. Keys expire after a day or more.

```sql
CREATE TABLE idempotency_keys (
  user_id       BIGINT NOT NULL,
  key           TEXT NOT NULL,
  request_hash  TEXT NOT NULL,
  status        TEXT NOT NULL,         /* 'in_progress', 'completed' */
  response_code INT,
  response_body JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, key)
);
```

### 5. State machines for payments and orders

Money objects move through defined stages, and not every jump is legal:

```
Order:    pending_payment -> paid -> fulfilled
                 |            |
                 v            v
              expired      refunded (full or partial)
Payment:  created -> authorized -> captured -> refunded
                \-> failed
```

Encode the allowed transitions in one function and enforce them with conditional updates: `UPDATE orders SET status='paid' WHERE id=:id AND status='pending_payment'`. If zero rows change, the order was not in the expected state, and you stop. This one habit prevents double fulfillment, refunding unpaid orders, and many "how did this happen?" incidents.

### 6. Integrating a payment provider

**Never let raw card numbers touch your server.** Use the provider's hosted checkout or hosted card fields. The card data goes from the browser straight to the provider, and you receive a token. This keeps you in the lightest tier of PCI DSS, the card industry's security standard. Use whichever provider supports your country. The pattern is the same everywhere.

The flow, with the principle that **the provider's webhook is the source of truth, not the browser redirect**:

```
1. Customer clicks Buy.
2. Your server, in ONE transaction: creates Order(status = pending_payment),
   reserves tickets for 10 minutes, creates PaymentAttempt(status = created) with an idempotency key.
3. Your server asks the provider to create a payment session. Returns the checkout URL to the browser.
4. Customer pays on the provider's page.
5. Provider sends a signed webhook "payment succeeded" to your server.
6. Your server verifies the signature, then in ONE transaction:
   marks the payment captured, marks the order paid, writes ledger entries,
   and writes an outbox event (Day 6).
7. A background worker sends the receipt and confirmation.
8. The browser, meanwhile, polls "GET /orders/:id" and shows the new status.
```

Why not trust the redirect back to your site? A user can close the tab before it loads. A user can also type the success URL by hand. Only the signed webhook, or a server to server API call to the provider, proves the payment happened.

### 7. Handling webhooks correctly

* **Verify the signature** (usually an HMAC of the body with a shared secret). Reject anything unsigned.
* **Respond fast** with 2xx and do heavy work asynchronously. Providers retry on timeouts.
* **Be idempotent.** Providers deliver "at least once," so duplicates are normal. Store each provider event id with a unique constraint and ignore repeats.
* **Expect disorder.** A "refund" event can arrive before the "capture" event. Use the state machine, and when in doubt fetch the current object from the provider API instead of trusting the payload order.
* **Keep a raw event log** of everything received. You will need it for debugging.

### 8. Reconciliation: trust, but verify

Servers crash. Webhooks get lost. Run a **reconciliation job** on a schedule: find payment attempts stuck in `created` for more than N minutes, ask the provider for their real status, and fix the records. Daily, compare your ledger totals with the provider's settlement report. Mismatches mean a bug, and finding them yourself is far better than hearing it from finance.

### 9. Refunds and other details

* A refund is a **new ledger transaction** that reverses the original movement. Never edit the original.
* Cap refunds: the sum of refunds for a payment cannot exceed what was captured. Enforce with a lock or constraint.
* Store all timestamps in **UTC**. Convert only for display.
* Never log full card data, tokens, or secrets.
* Every money affecting endpoint needs an idempotency key, authorization, an audit log entry, and a rate limit.

## Best practices checklist for money

* Integer minor units or `DECIMAL`, plus currency.
* Immutable double entry ledger. Entries per transaction sum to zero.
* Idempotency keys on every write that moves money.
* Explicit state machines with conditional updates.
* Hosted card fields. Signed, deduplicated, fast webhooks.
* Reconciliation jobs and alerts on mismatches.
* Concurrency tests with invariants.

## Build task for Day 5

Build the wallet: accounts, ledger, a "top up" endpoint backed by a **fake payment provider** you write yourself (a small separate service that calls your webhook after 3 seconds), and a ticket purchase paid from the wallet. Add idempotency keys and test: double click (same key twice), 20 parallel purchases against a balance that only covers 5, and a replayed webhook.

## Research list for Day 6

1. Double entry bookkeeping basics. Read a public engineering blog post about how a payments company models its ledger.
2. Idempotency keys. Read a payment API's documentation section on it.
3. PCI DSS: why hosted fields shrink your compliance scope.
4. Webhook signature verification and retry behavior.
5. At-least-once versus exactly-once delivery.
6. The saga pattern for multi step business processes.

## Questions for Day 6

1. Why store a ledger instead of just a balance column? When is a cached balance still acceptable?
2. How does an idempotency key work end to end? What if two requests with the same key arrive at the same instant?
3. A payment succeeds but the user closes the browser before the redirect. What should your system do, and why is the redirect not trusted?
4. The provider delivers the same webhook twice, and a refund event arrives before the capture event. How do you cope?
5. Why can't you put "charge the card" and "update my database" into one transaction? What do you do instead?
6. The charge succeeded at the provider but your server crashed before saving the result. How do you find and fix it?

***

# Day 6: Doing Work Later: Queues and Background Jobs

## Homework answers from Day 5

### Q1. Ledger versus a balance column

A balance column is a number with no memory. When it is wrong (and eventually it will be), you cannot tell which operation broke it. A ledger records every movement, so you can explain any balance, audit any dispute, rebuild state after a bug, and prove correctness (entries sum to zero). A **cached** balance is still fine and useful for fast reads, provided three things: it is updated in the same transaction as the ledger entries, it has a `CHECK` constraint, and a job periodically recomputes it from the ledger and alerts on any difference. Think of it as a speed optimization on top of the truth, never the truth itself.

### Q2. Idempotency key, end to end

The client picks a UUID per logical action and sends it. The server tries to **insert** the key row, with a unique constraint on (user, key). Because the database serializes inserts, exactly one concurrent request wins. The winner processes the request and saves the response on the key row. Anyone else with the same key gets either the stored response (if finished) or a 409 "still in progress" (if not). A same key with a different payload is rejected, since it signals a client bug. The unique constraint is doing the real work, which echoes Day 1 question 6: concurrency problems are solved by the database, not by `if` statements.

**Research note:** Stripe's API popularized this pattern and documents it well. Typical retention is 24 hours or longer. Also remember the key must be generated *per user action*, not per HTTP attempt: a retry reuses the same key.

### Q3. Payment succeeded but the browser closed

Nothing special should be needed. The provider's signed webhook arrives at your server regardless of what the browser does, your server marks the order paid, and the user sees the correct status next time they open the app (or gets the confirmation notification). The redirect is untrusted because it is just a browser navigating to a URL: it can fail to happen, be repeated, or be forged by someone typing the success URL. The webhook (or a server to server status query) is the only trustworthy signal.

### Q4. Duplicate and out of order webhooks

Duplicates: keep a table of processed provider event ids with a unique constraint, and process each event inside a transaction that first inserts its id. A repeat fails the insert and is ignored. Ordering: do not assume arrival order equals real order. Use the state machine so an event that does not fit the current state is not applied blindly. Better, treat the webhook as a **nudge**: when it arrives, fetch the latest state of the object from the provider and reconcile to that. If a refund arrives for a payment you have not yet marked captured, park the event and retry shortly, or apply the fetched final state.

### Q5. Why one transaction cannot cover the card charge and your database

A database transaction can only roll back what the database did. It cannot unsend a request to another company's servers. If you charge, then your commit fails, the customer was charged for nothing. If you commit, then the charge fails, you have an unpaid order marked paid. So you do it in steps with **durable intent**: record `PaymentAttempt(created)` with an idempotency key first, call the provider with the same key, record the result when known, and rely on webhooks plus reconciliation to settle uncertainty. Failures are handled by **compensating actions** (void or refund), not by rollback. The multi step version of this is called a **saga**.

### Q6. Charged, then server crashed

That is exactly why we wrote the `PaymentAttempt` row *before* calling the provider. After the crash, a reconciliation job finds attempts that have sat in `created` longer than some threshold, queries the provider using the idempotency key or metadata (the attempt id we attached), and updates the records. Often the webhook arrives anyway and fixes it first. Either way, the data to repair the situation exists, which is the whole point of writing intent before action.

***

## Today's lesson

### 1. Why queues exist

A restaurant waiter does not cook at your table while you wait. They write a slip and hand it to the kitchen, then keep serving. The slip sits on a rail, and cooks take slips as they become free. That rail is a queue.

Use background jobs when work is:

* **Slow:** generating a PDF ticket, resizing an image.
* **Unreliable:** calling the email provider or the payment provider.
* **Bursty:** 10,000 people hit "buy" the second tickets go on sale. The queue absorbs the spike while workers process at a steady pace.
* **Scheduled:** "remind the customer the day before the event," "release unpaid holds after 10 minutes."
* **Not the user's business:** the buyer should not wait for the receipt email to be sent.

### 2. The parts

```
Producer (your API)  ->  Queue / broker  ->  Consumer (worker)  ->  does the job
                                              |  acknowledges success
                                              |  or fails and the job is retried
```

* **Job:** a small message: a job type and an identifier, such as `send_receipt` with `order_id = 812`.
* **Broker:** the system holding jobs. Choices:

| Option | Good for | Notes |
| - | - | - |
| Redis based (BullMQ, Sidekiq style, Laravel queues, Asynq) | Most web app background jobs | Fast and simple. Check persistence settings |
| RabbitMQ | Rich routing, many job types | Mature message broker |
| Cloud queues (Amazon SQS, Google Cloud Tasks, Pub/Sub) | Managed, scales without ops | Pay per use, vendor specific |
| Kafka | High volume event streams, replay | A log, not a simple job queue. Heavier to run |
| Your own database (SKIP LOCKED, pg-boss, River) | Small to medium, want transactional enqueue | One less system, and enqueue can be in the same transaction as your data |

For TicketBox I would start with a database backed queue or Redis. Do not run Kafka for a project this size.

### 3. The delivery guarantee you must understand

* **At most once:** a job may be lost, never duplicated.
* **At least once:** a job is never lost, but may run **more than once**.
* **Exactly once:** across independent systems, essentially a myth. What you can build is *effectively once*: at least once delivery plus **idempotent jobs**.

Why do duplicates happen on a healthy broker? The worker finishes the job, then crashes before acknowledging. Or the job runs longer than its visibility timeout and the broker hands it to another worker. Or the network drops the acknowledgement. The broker cannot tell success from failure, so it safely retries.

**Therefore every job must be safe to run twice.** Techniques: a unique key per effect (`receipt:order_812`) with a unique constraint; check "already sent?" before sending; use conditional state changes (`WHERE status='held'`).

### 4. Retries done properly

* **Exponential backoff:** wait 10 seconds, then 20, 40, 80... so you do not hammer a struggling dependency.
* **Add jitter:** randomize each wait a little. Without jitter, 5,000 jobs that failed together all retry at the same instant and knock the recovering service over again (a thundering herd).
* **Set a maximum number of attempts.**
* **Separate retryable from permanent failures.** A timeout, a 503, a deadlock: retry. A 400 validation error, a missing required record, a bug: retrying will never help. Send these to a dead letter queue straight away.
* **Dead letter queue (DLQ):** where jobs go after exhausting retries. It is a hospital waiting room for sick jobs. Alert when it is non empty, investigate, fix, and **replay**.

### 5. Send identifiers, not whole objects

Put `order_id` in the job, not a copy of the order. By the time the job runs, the order may have changed (refunded, edited). Reading fresh data at run time avoids acting on stale facts, keeps payloads small, avoids leaking personal data into the broker, and survives changes to the data shape. The job should also **check whether it is still relevant** when it starts.

### 6. The dual write problem and the outbox pattern

Your API does two things: (1) saves the order in the database and (2) puts a `send_receipt` job on the queue. Two systems, two writes. If the server crashes between them, you either have an order with no receipt job, or (if you reverse the order) a job for an order that rolled back. There is no atomic way to write to two systems.

The fix is the **outbox pattern**:

1. In the **same database transaction** as the order, insert a row into an `outbox` table: "event: OrderPaid, order 812."
2. A separate **relay** process reads unsent outbox rows, publishes them to the queue, and marks them sent.
3. If the relay crashes after publishing but before marking, it will publish again. That is fine, because consumers are idempotent.

You get a guarantee that a committed business change always leads to its event, at least once.

A related trap: if you enqueue a job *inside* your transaction, the worker may pick it up **before the commit is visible** and find nothing. Enqueue after commit, or use the outbox.

### 7. Delayed and scheduled jobs

Two typical needs in TicketBox:

* **Hold expiry:** when a customer reserves seats, schedule a job for 10 minutes later: "if this hold is still unpaid, release the seats." Use delayed jobs for event driven timers.
* **Reminders and nightly work:** a scheduler (cron, or your queue library's scheduler) creates jobs: "find events starting tomorrow and enqueue reminders." Make sure only **one** instance of the scheduler fires a given run, or use idempotent job keys so duplicates collapse.

**Timers fire late and sometimes wrongly.** When the hold expiry job runs, the customer might have paid one second earlier. So the job must re-check state *atomically*:

```sql
UPDATE reservations SET status = 'expired'
WHERE id = :id AND status = 'held';
/* if 0 rows changed, it was paid or already handled. Do nothing. */
```

### 8. Operating queues in the real world

* **Separate queues** for critical work (payments, security emails) and bulk work (marketing). A flood of bulk jobs must never delay a password reset.
* **Concurrency limits** per queue, to protect downstream APIs and your database.
* **Ordering** is not guaranteed with many workers. If order matters per entity, route by key (all jobs for order 812 to the same partition) or design jobs so order does not matter.
* **Graceful shutdown:** on deploy, workers should finish the current job, stop taking new ones, and exit cleanly.
* **Monitor:** queue depth, job age (time waiting), failure rate, DLQ size, retry counts. A growing queue means workers cannot keep up.
* **Never lose the failure information:** log job id, attempt number, and error, with a correlation id that links back to the original request.

## Best practices checklist for queues

* Every job idempotent.
* Identifiers in payloads. Re-check state at run time.
* Backoff with jitter. Retry only transient errors. DLQ plus alerting plus replay.
* Outbox for "database change implies event."
* Separate queues by criticality.
* Timers re-check state atomically.
* Metrics and alerts on depth, age, failures.

## Build task for Day 6

Add a queue to TicketBox. Implement: `send_receipt(order_id)` as an idempotent job, make it fail randomly 50 percent of the time and watch the retries, force a permanent failure and watch it land in the DLQ, then replay it. Implement hold expiry with a delayed job and the atomic conditional update. Add an `outbox` table and a small relay.

## Research list for Day 7

1. At-least-once versus exactly-once. Kafka versus RabbitMQ versus SQS versus Redis queues.
2. The outbox pattern and transactional messaging.
3. Exponential backoff with jitter (the AWS Architecture Blog has a clear article).
4. Dead letter queues and poison messages.
5. The competing consumers pattern.
6. Cron versus delayed jobs, and distributed locks.

## Questions for Day 7

1. Why can the same job run twice even on a healthy broker, and how do you make a job safe?
2. Why add jitter to retry delays?
3. What is the dual write problem, and how does the outbox pattern solve it?
4. Which errors should be retried, and which should go straight to the dead letter queue?
5. Why put an order ID in a job instead of the whole order object?
6. The hold expiry job fires, but the user paid one second ago. What stops you from releasing a paid seat?

***

# Day 7: Telling People What Happened: Notifications

## Homework answers from Day 6

### Q1. Why can a job run twice, and how do you make it safe?

On a healthy broker, duplicates still happen when a worker finishes the work but dies before acknowledging, when a job outlives its visibility timeout and gets handed to a second worker, or when the network loses the acknowledgement. The broker cannot distinguish "done" from "crashed," so it chooses the safe side and delivers again. Make jobs safe with a **natural dedupe key** (`receipt:order_812`) backed by a unique constraint, a check of "has this already been done?" at the start, and conditional state changes (`WHERE status = 'held'`). Think of a bank teller who checks the slip number before paying out: the same slip presented twice pays once.

### Q2. Why add jitter?

Failures are usually correlated. If the email provider is down for a minute, thousands of jobs fail at nearly the same moment. With plain exponential backoff they all retry at exactly 10 seconds, then exactly 20, and so on, arriving as synchronized waves that can knock the recovering service over again. Jitter adds a random offset to each wait so retries spread out smoothly. **Research note:** the AWS Architecture Blog compares "full jitter" and other variants. Full jitter (random wait between zero and the backoff cap) is a good default.

### Q3. Dual write and the outbox

Writing to the database and to the queue are two separate systems, and no mechanism can make both succeed or both fail. A crash between them either loses the job or produces a job about data that was rolled back. The outbox pattern turns the second write into a local database write: the event goes into an `outbox` table in the **same transaction** as the business change, so they commit together. A relay then publishes outbox rows to the broker and marks them sent. If the relay publishes and crashes before marking, it publishes again, which is acceptable because consumers are idempotent. More advanced setups read the database's change log (change data capture, for example with Debezium) instead of polling the table.

### Q4. Retry or dead letter?

**Retry:** timeouts, connection failures, 5xx responses, 429 rate limits (respect the `Retry-After` header), deadlocks and serialization failures. These are temporary. **Straight to the DLQ:** validation errors, 4xx responses that say the request is wrong, a required record that does not exist, permission failures, and exceptions that look like bugs. Retrying a bug 10 times just makes 10 failures. After a fix is deployed, replay the DLQ.

### Q5. Why an ID instead of the whole object?

The data may change between enqueue and execution: the order could be refunded, or the user could change their email. A job holding a stale copy would act on facts that are no longer true. IDs also keep messages small, keep personal data out of the broker and its logs, and make it easy to evolve your data model without breaking old queued jobs. The job loads fresh data and checks "is this still relevant?"

### Q6. Hold expiry firing just after payment

The expiry job must not decide from what it believed at scheduling time. It must change state with an atomic conditional update: `UPDATE reservations SET status='expired' WHERE id=:id AND status='held'`. If the payment webhook already moved the reservation to `paid`, the condition fails, zero rows change, and the job does nothing. The same row has a single status, and the database serializes the two competing updates, so one of them wins cleanly. If it matters, the loser also learns it lost by checking the affected row count.

***

## Today's lesson

### 1. Two very different kinds of messages

* **Transactional:** the user needs it because of something they did. Receipts, password resets, security alerts, ticket confirmations. These are expected and legally simpler.
* **Marketing and promotional:** "check out this new event." These require consent and an easy way to unsubscribe in most places (laws such as CAN-SPAM in the United States and GDPR in Europe set rules). Keep them on separate infrastructure so a marketing complaint cannot hurt delivery of password resets.

### 2. The channels

| Channel | Strengths | Watch out for |
| - | - | - |
| In-app (a bell icon with a list) | Free, reliable, you control it | Only seen when the app is open |
| Email | Universal, good for records and long content | Spam filters, deliverability, delays |
| Push (FCM for Android, APNs for iOS, web push) | Reaches the user when the app is closed | Permission prompts, device tokens expire |
| SMS | High open rate | Costs money, regulated, and the weakest channel for security codes |
| Webhooks, chat tools | Integration with other systems | Retries, signatures, security |

### 3. The architecture

Notification code spreads quickly through an application and becomes a mess. Keep it in one place:

```
Business code finishes work and records an EVENT:  "OrderPaid(order 812)"   (via the outbox)
        |
        v
Notification service (a queue consumer):
   1. Who should be told?   (the buyer, the organizer)
   2. What do they prefer?  (email on, push off, quiet hours)
   3. Build the message from a template and the user's language
   4. Create one job per channel
        |
        v
Channel workers -> provider adapter (email, push, SMS) -> record the result
```

Notice that your payment code does not know how emails are sent. It just says "the order was paid." This is the difference between an **event** (something happened, past tense, any number of listeners) and a **command** (please do this, one handler).

### 4. Data model

```sql
CREATE TABLE notifications (          /* the in-app inbox */
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES users(id),
  type        TEXT NOT NULL,           /* 'order_paid', 'event_reminder' */
  payload     JSONB NOT NULL,          /* ids and small facts, rendered at display time */
  read_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notification_preferences (
  user_id BIGINT NOT NULL, type TEXT NOT NULL, channel TEXT NOT NULL, enabled BOOLEAN NOT NULL DEFAULT true,
  PRIMARY KEY (user_id, type, channel)
);

CREATE TABLE device_tokens (
  id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL, platform TEXT NOT NULL,
  token TEXT NOT NULL UNIQUE, last_seen_at TIMESTAMPTZ, invalid_at TIMESTAMPTZ
);

CREATE TABLE notification_deliveries (
  id BIGSERIAL PRIMARY KEY,
  dedupe_key TEXT NOT NULL UNIQUE,     /* e.g. 'order_paid:812:user_5:email' */
  channel TEXT NOT NULL, provider_message_id TEXT,
  status TEXT NOT NULL,                /* queued, sent, delivered, bounced, failed */
  error TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

The `dedupe_key` with a unique constraint is our Day 6 idempotency idea applied here: the same event for the same user on the same channel can be sent only once.

### 5. Email: the part everyone underestimates

Sending an email is easy. Getting it into the inbox is the hard part.

* **Never send from your API request.** Always a queue job (Day 6). The provider can be slow or down, and a slow email must not slow down or fail a purchase.
* **Use a provider** (a transactional email service) instead of sending directly from your own server. They manage reputation and delivery.
* **SPF, DKIM, DMARC:** three DNS records that prove your emails are really from you. SPF lists which servers may send for your domain. DKIM signs each message. DMARC tells receivers what to do when checks fail. Without them, mail lands in spam or is rejected.
* **Bounces and complaints.** A hard bounce (address does not exist) means stop sending to that address. A spam complaint means stop sending marketing mail immediately. Providers report both through webhooks. Keep a **suppression list**. Ignoring bounces damages your sender reputation, and then even your important mail goes to spam.
* **Templates:** keep presentation separate from logic, always send a plain text version alongside HTML, and support the user's language. Preview templates in development.
* **Links:** use signed, expiring links for anything sensitive. Never include secrets in an email body.

### 6. Push and SMS

* Push needs a **device token** per installed app. Tokens change and expire. When the push service says a token is invalid, mark it dead and stop using it.
* Do not put private details on a lock screen preview. "You have a new notification" can be the safe default for sensitive content.
* SMS is expensive and tightly regulated in many countries. For security codes, prefer an authenticator app or passkeys, because SMS can be intercepted or hijacked through SIM swapping.

### 7. Real time in-app notifications

How does the bell icon update without a page refresh?

* **Polling:** the frontend asks "anything new?" every 30 seconds. Simplest, works everywhere, wastes some requests.
* **Server-Sent Events (SSE):** one long lived HTTP connection where the server pushes messages to the browser. One direction only, reconnects automatically, easy to run. A very good fit for notifications.
* **WebSockets:** two way, persistent. The right tool for chat and collaborative editing, but heavier. With multiple servers you also need a shared channel (such as Redis pub/sub) so an event reaches the server holding that user's connection.

My advice for TicketBox: start with polling, move to SSE when you care about speed, and use push notifications for when the app is closed.

### 8. Timing, volume, and respect

* **Time zones:** "remind me at 9 am" means 9 am *their* time. Store the user's IANA time zone name (for example `Europe/Paris`), not a numeric offset, because daylight saving changes offsets. Store event times in UTC and compute the send time with a time zone library.
* **Rate limit and batch.** Ten notifications in a minute is spam. Combine them into a digest.
* **Quiet hours** and per-type preferences, with one click unsubscribe.
* **Notification fatigue is real.** Every message must earn its place. A user who mutes everything is lost.

### 9. Provider abstraction, testing, and observability

* Hide each provider behind a **small interface** (`send(message)`), so you can switch providers, add a fallback during an outage, and use a fake in tests. Keep it small. A huge abstraction that tries to cover every provider's features becomes a burden.
* **In development,** use a local mail catcher (Mailpit is one) so no real email leaves your machine. **In staging,** restrict sending to an allowlist of test addresses. Many companies have accidentally emailed real customers from test environments.
* **Measure:** delivery rate, bounce rate, complaint rate, time from event to delivery, queue depth, and provider error rates.

## Best practices checklist for notifications

* Event in, notification out. Business code never calls email directly.
* Every send goes through a queue and is idempotent (dedupe key).
* Preferences, suppression lists, and unsubscribe handled properly.
* SPF, DKIM, DMARC configured. Bounces and complaints processed.
* Provider behind a small interface. Fakes in tests. Allowlist in staging.
* Time zones stored as IANA names.
* Sensitive content kept out of previews.

## Build task for Day 7

Build the notification service for TicketBox: an in-app inbox with unread count, `order_paid` and `event_reminder` types, preferences, a deliveries table with dedupe keys, and an email sender behind an interface (use a mail catcher). Make the email adapter fail sometimes and confirm retries do not send duplicates. Add a reminder scheduled for 9 am in the user's own time zone, the day before the event.

## Research list for Day 8

1. SPF, DKIM, and DMARC: what each record does.
2. How FCM and APNs deliver push notifications and manage device tokens.
3. WebSockets versus Server-Sent Events versus polling.
4. Transactional versus marketing email, and the basics of unsubscribe and consent rules.
5. "Design a notification system": read two public write ups of how large products do it.
6. Event driven architecture: events versus commands.

## Questions for Day 8

1. Why must email be sent from a background job and never from the HTTP request?
2. How do you avoid sending the same notification twice?
3. A user's email address keeps bouncing. What should the system do?
4. Polling, SSE, or WebSockets for in-app notifications: how do you choose?
5. A user wants a reminder at 9 am local time. How do you design that, including daylight saving?
6. Why put the email provider behind an interface, and what is the risk of overdoing it?
7. What is the difference between an event and a command?

***

# Day 8: Putting It All Together: TicketBox End to End

## Homework answers from Day 7

### Q1. Why email from a background job?

Because the email provider is a third party you do not control. It can be slow, rate limiting you, or down. If you send inside the request, the user waits for it, a provider outage turns into failed purchases, your database connection sits idle waiting on the network, and a failure of a *nice to have* email wrongly fails a *must have* payment. In a job, the purchase completes immediately, the email retries safely with backoff, and failures are visible in the queue metrics.

### Q2. Avoiding duplicate notifications

Assume the job will run twice, and make the second run harmless. Build a deterministic dedupe key from (event id, user, channel), insert it into the deliveries table with a unique constraint *before* sending, and skip the send if the insert fails. There is a small gap: crash after sending but before recording "sent" can cause one duplicate on retry. Where providers support their own idempotency key, pass it along. In practice, a rare duplicate email is better than a lost receipt, which is why systems favor at least once delivery.

### Q3. An address that keeps bouncing

A hard bounce means the address is invalid, so add it to the suppression list and stop sending. A soft bounce (mailbox full, temporary server issue) deserves a few retries, then treat it as failing. Surface the problem to the user in the app ("we could not reach your email, please update it") through the in-app channel, which still works. For a spam complaint, stop marketing mail at once. This protects your sender reputation, which is a shared, fragile asset that determines whether *any* of your mail reaches inboxes.

### Q4. Polling, SSE, or WebSockets?

Choose by how real time it must be and which direction data flows. Polling is the simplest, perfectly fine for low frequency updates, and the easiest to scale. SSE is one directional server to browser over normal HTTP, reconnects by itself, and is the neat fit for notifications and live status. WebSockets are bidirectional, designed for chat, multiplayer, and collaboration, but they need extra infrastructure at scale (sticky routing or a shared pub/sub). For the closed app case, none of them helps: you use push notifications.

### Q5. A reminder at 9 am local time

Store the event start in UTC and the user's time zone as an IANA name. Compute the reminder moment with a time zone library: "the day before the event, 09:00 in `Europe/Paris`," converted to UTC. Schedule a delayed job for that instant, keyed `reminder:event_9:user_5` so duplicates collapse. Daylight saving is handled automatically *because* you used a named zone instead of a fixed offset. Recompute (and cancel the old job) if the user changes time zone or the event is rescheduled. When the job runs, re-check that the ticket is still valid and the event still exists.

### Q6. Why an email interface, and the risk of overdoing it

Benefits: you can change providers, add a fallback provider for outages, test with a fake, and keep logging consistent. The risk: designing a giant "universal" interface for imaginary future providers, which hides useful provider features and adds complexity you never use. Keep it tiny: `send(to, subject, html, text, idempotencyKey)` and a result. Add things only when a real need appears.

### Q7. Events versus commands

An **event** states a fact that already happened: `OrderPaid`. The sender does not care who listens, and there can be zero, one, or many listeners (receipt, analytics, organizer alert). A **command** is a request addressed to one handler: `SendReceipt`. Events keep systems loosely coupled, since adding a new reaction means adding a listener, not editing the payment code. Commands are clearer when you need exactly one thing to happen.

***

## Today's lesson: designing TicketBox properly

This is a design review, the kind I do with my team before anyone writes code. We combine all seven days.

### 1. Roles and permissions

| Role | Key permissions |
| - | - |
| customer | event.view, order.create, order.view_own, wallet.topup, order.request_refund |
| organizer | everything a customer has, plus event.create, event.edit_own, sales.view_own, order.refund_own_event |
| admin | all of the above on any resource, plus user.manage, order.refund_any, audit.view |

Objects are protected by ownership rules on top of this (Day 3).

### 2. Core tables

* `users`, `refresh_tokens`, `password_resets`, `email_verifications`
* `roles`, `permissions`, `role_permissions`, `user_roles`
* `events` (organizer id, starts_at in UTC, status), `ticket_types` (event id, price in minor units, currency, remaining, version)
* `reservations` (user, ticket type, quantity, status: held, paid, expired, expires_at)
* `orders`, `order_items` (status state machine)
* `payment_attempts` (order id, provider, provider reference, idempotency key, status)
* `accounts`, `ledger_transactions`, `ledger_entries`
* `idempotency_keys`, `webhook_events` (provider event id unique)
* `outbox`, `jobs` (if the queue is in the database)
* `notifications`, `notification_preferences`, `notification_deliveries`, `device_tokens`
* `audit_log` (append only)

### 3. API surface (selected)

| Endpoint | Auth | Notes |
| - | - | - |
| POST /auth/register, /auth/login, /auth/refresh, /auth/logout | public, then token | rate limited |
| GET /events, GET /events/:id | public | pagination, filters |
| POST /events | organizer | allowlisted fields, audit |
| PATCH /events/:id | owner organizer or admin | optimistic lock (`version`) |
| POST /reservations | customer | atomic inventory decrement, schedules hold expiry |
| POST /orders | customer | idempotency key required |
| POST /payments/webhook | provider only | signature verified, deduplicated, fast |
| POST /orders/:id/refund | permission plus ownership | idempotency key, ledger reversal |
| GET /notifications, POST /notifications/:id/read | customer | scoped to current user |

### 4. The purchase flow, step by step

Read this and for every step ask: **what happens if the server dies right after it?** That question is the heart of the craft.

1. **Request arrives.** Rate limit, authenticate, check permission `order.create`, validate the body with an allowlist. Require an `Idempotency-Key`.
2. **Claim the idempotency key** (unique insert). A duplicate returns the stored response.
3. **One database transaction:**
   * `UPDATE ticket_types SET remaining = remaining - :n WHERE id = :id AND remaining >= :n` (zero rows means sold out, so stop).
   * Insert the reservation with `status = 'held'`, `expires_at = now() + 10 minutes`.
   * Insert the order as `pending_payment`, and a `payment_attempt` as `created`.
   * Insert outbox events: `HoldCreated`.
   * Commit.
4. **After commit,** the relay publishes `HoldCreated`, and a delayed job `expire_hold(reservation_id)` is scheduled for 10 minutes.
5. **Create the provider payment session** using the attempt id as idempotency key. Return the checkout URL. (If wallet paid, do the wallet ledger move in step 3 instead and skip the provider.)
6. **Customer pays.** The provider sends the signed webhook.
7. **Webhook handler:** verify signature, insert the provider event id (duplicates stop here), then in one transaction: payment `captured`, order `paid` (conditional on `pending_payment`), reservation `paid` (conditional on `held`), ledger entries (sum zero), outbox `OrderPaid`. Respond 2xx.
8. **Notification consumer** handles `OrderPaid`: builds receipt and confirmation for the buyer, a sales alert for the organizer, with dedupe keys. Jobs retry with backoff and jitter.
9. **Reminder job** scheduled for the day before the event at 9 am in the buyer's time zone.
10. **If unpaid at 10 minutes,** `expire_hold` runs the atomic conditional update, returns the tickets to inventory, and marks the order expired.
11. **Reconciliation job** every few minutes: payments stuck in `created` are checked against the provider.

Every step has a clear answer to the crash question because of three habits: intent is written before action, state changes are conditional, and all side effects are idempotent.

### 5. A failure mode table

Keep one of these for any serious feature. It is worth more than the code.

| Failure | Result | Defense |
| - | - | - |
| Two users buy the last ticket | One wins | Conditional atomic update |
| User double clicks Pay | One order | Idempotency key |
| Server crashes after charge, before save | State unknown | Payment attempt written first, webhook, reconciliation |
| Webhook delivered twice | One effect | Event id unique constraint |
| Webhook arrives out of order | No illegal state | State machine, fetch latest from provider |
| Email provider is down | Receipt delayed, not lost | Queue, retries, DLQ |
| Job runs twice | One email | Dedupe key |
| Hold expiry fires after payment | Paid seat kept | Conditional update on `status = 'held'` |
| Token stolen | Contained | Short access token, rotation, reuse detection |
| Customer reads another order | Denied | Ownership in query, matrix tests |

### 6. Build order I would use

1. Auth (Days 1 and 2), with tests.
2. Roles, permissions, policies, matrix tests (Day 3).
3. Events and ticket types, the atomic purchase and concurrency test (Day 4).
4. Wallet, ledger, idempotency, fake provider, webhooks (Day 5).
5. Queue, outbox, hold expiry, retries, DLQ (Day 6).
6. Notifications and reminders (Day 7).
7. Observability, hardening, deployment.

### 7. Testing strategy

* **Unit tests** for pure logic (state machine, fee rounding, policies).
* **Integration tests against a real database,** not mocks, for transactions and constraints.
* **Concurrency tests** with invariants.
* **Authorization matrix tests.**
* **Webhook tests:** replay, duplicate, out of order, bad signature.
* **Job tests:** run the same job twice, kill it midway, force failures.
* **A "money invariant" test** that runs after every scenario: ledger entries sum to zero per transaction, no negative wallets, tickets sold never exceed capacity.

### 8. Observability

* **Structured logs** (JSON) with a **request id** that follows the request into jobs, so one purchase can be traced across API, webhook, and worker.
* **Metrics:** request rate and latency, error rate, queue depth and age, DLQ size, payment success rate, webhook lag, bounce rate.
* **Alerts** on symptoms users feel (purchase failures, growing queue age, stuck payments) more than on raw CPU.
* **Error tracking** and a **health check** endpoint.
* Distributed tracing (OpenTelemetry) becomes valuable when you have several services.

### 9. Production readiness checklist

* HTTPS only. Security headers. A strict CORS allowlist.
* Configuration from environment variables. Secrets in a secrets manager, never in the repository.
* Database migrations that are **backward compatible** during deploys (add a new column, deploy code that writes both, backfill, switch reads, then drop the old one). A direct rename breaks running servers.
* Backups, **and a tested restore.** A backup you never restored is a hope.
* Pagination on every list endpoint. Request size limits. Input validation everywhere.
* Graceful shutdown for the API and workers. Health and readiness checks.
* Dependency updates and vulnerability scanning.
* Feature flags for risky changes. A rollback plan for every deploy.

### 10. Scaling, in the right order

Do not design for a million users on day one, but know the order of moves when growth arrives: fix slow queries and add indexes, remove N+1 queries, cache hot reads, add read replicas, scale workers horizontally, split out services only when a team or a load pattern demands it. Most systems that fail early fail from correctness bugs and missing observability, not from lack of scale.

### 11. Habits of a good engineer

* Write a **one page design note** before building anything with money or permissions. Ask a colleague to attack it.
* For every step, ask: "what if this fails, runs twice, or runs late?"
* Enforce invariants in the database.
* Prefer boring, proven tools.
* Make the unsafe thing hard to do by accident (deny by default, constraints, small interfaces).

## Final build task

Complete TicketBox. Then run a **game day**: kill the worker mid job, send duplicate webhooks, fire 200 concurrent purchases, expire holds while payments land, revoke a role mid session. Write down what broke and fix it. Keep that write up. It is a better interview story than any tutorial project.

## Research list (final)

1. *Designing Data-Intensive Applications* by Martin Kleppmann: chapters on transactions and on messaging.
2. The free online *Site Reliability Engineering* book: chapters on monitoring and handling overload.
3. The Twelve-Factor App methodology.
4. OpenTelemetry basics: logs, metrics, traces.
5. Zero downtime database migrations.
6. Caching: where it helps, and why invalidation is hard.

## Final questions (answers in Day 9)

1. In the purchase flow, what happens if the server crashes right after each of steps 3, 5, 7, and 8?
2. Which operations in TicketBox need idempotency keys, and which do not?
3. How would you deploy a column rename with zero downtime?
4. What would you monitor and alert on, and why those things?
5. Traffic grows 100 times. What breaks first, and in what order do you respond?
6. Where does authorization live in the codebase, and how do you make sure nothing bypasses it?

***

# Day 9: Final Answers and What Comes Next

## Answers to the Day 8 questions

### Q1. Crash after each step

* **After step 3 (transaction committed):** reservation, order, attempt, and outbox rows exist. The relay will publish `HoldCreated` on restart, and a reconciliation or the stored delayed job releases the hold if nobody pays. Nothing is lost, nothing is double counted. If the crash happened *before* commit, everything rolled back and the client's retry with the same idempotency key starts fresh.
* **After step 5 (provider session created):** the attempt record already exists with its idempotency key, so a retry reuses the same provider call and does not create a second payment session. The customer can still pay, and the webhook will find the attempt.
* **After step 7 (webhook transaction committed, response not yet sent):** the provider will retry the webhook because it saw no 2xx. The event id unique constraint turns the retry into a harmless no-op. Because the outbox row was committed with the state change, `OrderPaid` will still be published.
* **After step 8 (notification job ran, not acknowledged):** the broker redelivers the job and the dedupe key prevents a second email. The worst realistic case is a single duplicate if the crash fell between sending and recording.

The pattern: every durable decision is made in a transaction, and every side effect follows from a record that survives the crash.

### Q2. Which operations need idempotency keys?

Those that create or change something in a way that must not repeat and that a client might retry: creating an order, paying, topping up the wallet, refunding, and anything that sends money or sends a notification triggered by the client. Plain reads do not need one. Updates that set a field to a specific value (`PATCH name = "Sam"`) are naturally idempotent. Operations like "add 1 to a counter" are not, so they need protection. Webhooks are handled by the provider's event id instead of a client generated key.

### Q3. Column rename with zero downtime

Never rename in one step, because old servers still running during the deploy expect the old name. Use **expand and contract**: (1) add the new column; (2) deploy code that writes to both columns and reads the old; (3) backfill old rows in small batches; (4) deploy code that reads the new column; (5) later, deploy code that stops writing the old column; (6) drop the old column. Each step is safely reversible, and every version of the code that is running at any moment works with the schema as it is at that moment.

### Q4. What to monitor and alert on

Alert on **symptoms** that hurt users or money: purchase success rate dropping, payments stuck in `created`, queue age growing, DLQ not empty, webhook lag rising, ledger invariant check failing, elevated 5xx rate, bounce rate climbing, login failure spikes. Also dashboard the resources (CPU, memory, connection pool, disk), but page people mainly for symptoms. A good alert tells you what is broken and where to look first, and fires rarely enough that people trust it.

### Q5. 100 times the traffic

Typically the database breaks first: hot rows (the last tickets, popular event rows), slow queries, connection exhaustion. Response order: look at metrics to find the real bottleneck instead of guessing; add or fix indexes and remove N+1 queries; put a connection pooler in front of the database; cache hot read endpoints (event listings); move heavy work to the queue and scale workers; add read replicas for read heavy paths; for the on-sale spike use a **virtual waiting room** or a queue in front of the purchase endpoint so demand is admitted at a rate the system can handle; and only then consider partitioning data or splitting services.

### Q6. Where authorization lives

In one layer: policies, called through one function in the service layer (or a middleware that every route must declare), with **deny by default**. A test walks the route table and fails if any route lacks an explicit authorization declaration. Object level checks live inside queries (`WHERE id = :id AND user_id = :me`). Dangerous actions re-read permissions from the database. And the matrix test suite protects against regressions. Nothing in the frontend counts as enforcement.

***

## Closing notes from me

You came into this session able to build a CRUD API. What separates that from a real application is not more code. It is a way of thinking, and it fits in a few sentences:

* **Identity** must be proven carefully and revoked easily.
* **Permission** is checked on the server, on every object, by default denied.
* **Truth about money** lives in an append only ledger, changed only through transactions, and protected against repetition by idempotency keys.
* **Anything slow or unreliable** goes through a queue, and every job can safely run twice.
* **Every step** gets the question: "what if this crashes here, runs twice, or runs late?"
* **The database** is your last line of defense, so put your invariants there.

### How I would use the next month

1. Build TicketBox in the order given. Do not skip the concurrency tests.
2. Run the game day and write it up.
3. Read the Kleppmann chapters on transactions and messaging.
4. Put the project, the design note, the failure mode table, and the game day write up in your repository. That combination shows judgment, which is what employers cannot get from a tutorial project.

Go and build it. When it breaks, and it will, that is the part where you learn the most.

***
