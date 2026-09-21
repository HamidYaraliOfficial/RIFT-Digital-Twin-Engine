# RIFT — Digital Twin Engine

**A universal, extensible platform for building, running, simulating, monitoring and analyzing Digital Twins of factories, buildings, data centers, warehouses, logistics hubs, urban infrastructure, and energy networks.**

Core engine in **Go** · Control Center in **TypeScript, React & Next.js** · Fully working, dependency-light reference implementation.

This document is written in three languages, each complete on its own:

- [🇬🇧 English](#-english)
- [🇮🇷 فارسی](#-فارسی)
- [🇨🇳 中文](#-中文)

---

## 🇬🇧 English

### 1. What is RIFT?

RIFT turns a real-world environment — a factory floor, an office tower, a data center, a warehouse, or a whole city district — into a **live digital twin**: a structured model of every entity, sensor, actuator, and relationship in that environment, kept in sync with real telemetry, capable of simulating its own future, and able to answer "what happens if…" before anything changes in the real world.

The project ships as two parts:

| Component | Stack | Responsibility |
|---|---|---|
| `backend/` | Go (standard library only) | Digital Twin Core, Entity/Relationship graph, telemetry ingestion, state engine, event engine, rules engine, simulation & What-If engine, analytics, security, REST + SSE API, CLI |
| `frontend/` | Next.js 14 · React 18 · TypeScript · Tailwind CSS | Control Center: dashboards, 3D/2D twin viewers, telemetry explorer, rule & scenario builders, operating-hours configurator, alerts, audit trail, permissions |

### 2. Feature Overview

**Digital Twin Core**
- Universal Entity model (type, properties, state, position/orientation/dimensions, metadata, owner, parent/children, health, status, version, timestamps)
- Typed Relationships (`contains`, `connected_to`, `powered_by`, `depends_on`, `located_in`, `controlled_by`, `communicates_with`, `feeds`, `produces`, `consumes`, `transports`, `monitors`) with confidence and history
- Twin Registry with create / clone / archive / snapshot / export / import
- Bundled templates: **Factory, Smart Building, Data Center, Warehouse, Smart City** — all fully customizable

**Telemetry & State**
- Sensor Abstraction Layer (temperature, humidity, pressure, vibration, energy, voltage, current, air quality, light, motion, speed, location, CPU, memory, disk, network, and custom types)
- Goroutine worker-pool ingestion pipeline with backpressure handling, quality checks, and gap/offline detection
- Four-way State Engine: Current / Previous / Expected / Simulated
- Time-series store with windowing and bucketed downsampling

**Automation & Intelligence**
- Rules Engine (trigger → conditions → actions, priority, cooldown, safety guard)
- Statistical Anomaly Detection (self-adjusting per-sensor baseline)
- Health / Risk / Criticality scoring per entity and per twin
- Dependency Graph Engine (impact analysis, shortest path, reachability, dependency depth)
- Root Cause Analysis (event correlation across the dependency graph)
- Alerting Engine (severity, deduplication, cooldown, escalation, acknowledgement, webhook notifications)

**Simulation**
- Time Engine (pause / resume / step / seek / speed scaling, independent of wall-clock time)
- What-If Scenario Engine — every simulation runs on an isolated in-memory branch and **never touches live twin state**
- Fault Injection (sensor failure, network failure, power failure, equipment failure, communication delay, packet loss, resource exhaustion)
- Monte Carlo Engine and Parameter Sweep Engine
- Multi-objective Optimization Engine (combine cost, risk, throughput, energy objectives)

**Security & Multi-Tenancy**
- Session-based authentication (HMAC-signed tokens) and salted/stretched password storage
- RBAC + ABAC-style Twin Permission Model (role- or user-level, optionally scoped to a single entity)
- Command Safety Engine in front of every actuator command: allow-list, permission check, range check, interlocks, rate limiting, and an explicit confirmation step for sensitive commands
- Full audit trail persisted to disk

**Control Center (Frontend)**
- Control Tower dashboard with live KPIs and ingestion-pipeline throughput
- Entity tree + inspector, with one-click impact analysis and root-cause search
- Interactive **3D twin viewer** (Three.js / React Three Fiber) and a dependency-free **2D operational map**
- Real-time telemetry charts over Server-Sent Events (no extra transport dependency required)
- Rule builder, What-If scenario builder with Monte Carlo results, audit trail, and permission management
- **User-configurable Operating Hours**: you enter each day's opening/closing time yourself, and RIFT computes live open/closed status plus a countdown to the next change
- Five themes — **Windows (default), Light, Dark, Red, Blue** — and full **English / Persian / Chinese** localization, with automatic right-to-left layout for Persian

### 3. Requirements

- **Go 1.22 or newer** (backend has zero external dependencies — the standard library only)
- **Node.js 18.18 or newer** and **npm** (frontend)

### 4. Installation & Running

**Step 1 — Start the backend**

```bash
cd backend
go run ./cmd/riftd
```

The server listens on `http://localhost:8080` by default and creates a `data/` folder for its persisted state and audit log. On first launch it seeds one demo twin per bundled template and creates a default administrator account:

```
username: admin
password: admin123
```

Change this password immediately in any shared or production environment.

Optional flags / environment variables:

```bash
go run ./cmd/riftd -addr :8080 -data ./data
# or
RIFT_ADDR=:8080 RIFT_DATA_DIR=./data go run ./cmd/riftd
```

To build a standalone binary instead:

```bash
go build -o riftd ./cmd/riftd
./riftd
```

**Step 2 — Start the Control Center (frontend)**

In a second terminal:

```bash
cd frontend
npm install
cp .env.example .env.local   # points the UI at http://localhost:8080 by default
npm run dev
```

Open `http://localhost:3000` in your browser, sign in with the administrator account above, and you will see five ready-to-explore twins (Factory, Smart Building, Data Center, Warehouse, Smart City) already generating live simulated telemetry.

For a production build of the frontend:

```bash
npm run build
npm run start
```

**Step 3 — (Optional) Use the CLI**

```bash
cd backend
go build -o rift ./cmd/rift
export RIFT_SERVER=http://localhost:8080
./rift login --username=admin --password=admin123
export RIFT_TOKEN=<paste the token printed above>
./rift twin list
./rift twin create --name="My Plant" --template=factory
./rift scenario create <twinId> --name="Cooling failure" --ticks=40 --fault=equipment_failure:<entityId>:0:1
./rift scenario run <scenarioId>
./rift benchmark <twinId> --count=100000
```

Run `./rift help` for the full command reference (`twin`, `entity`, `sensor`, `ingest`, `simulate`, `scenario`, `snapshot`/`export`, `replay`, `analyze`, `optimize`, `benchmark`, `worker`).

### 5. Configuring Operating Hours

Open a twin's **Operating Hours** tab, give the schedule a label and time zone, and enter your own opening and closing time for every day of the week (or mark a day fully closed). RIFT immediately computes, and keeps live:

- whether the twin is open right now,
- what the next change will be (opening or closing), and
- exactly how much time is left until that change.

Nothing here is pre-filled with assumptions — every hour comes from what you type in.

### 6. Project Structure

```
RIFT-Digital-Twin-Engine/
├── backend/
│   ├── cmd/riftd/           # HTTP + SSE server entry point
│   ├── cmd/rift/            # CLI client
│   ├── internal/model/      # Universal Twin Model (entities, sensors, events, rules...)
│   ├── internal/registry/   # Twin Registry (graph store + persistence)
│   ├── internal/telemetry/  # Ingestion pipeline, adapters, time-series store
│   ├── internal/stateengine/ internal/events/ internal/rules/ internal/anomaly/
│   ├── internal/health/ internal/graph/ internal/rootcause/
│   ├── internal/simulation/ internal/optimization/ internal/actuator/
│   ├── internal/alerting/ internal/auth/ internal/audit/
│   ├── internal/operatinghours/ internal/synthetic/ internal/exportimport/
│   └── internal/api/        # REST + SSE handlers wiring every engine together
└── frontend/
    └── src/
        ├── app/              # Next.js App Router pages
        ├── components/       # Dashboard, 3D/2D viewers, panels, providers
        └── lib/              # API client, types, i18n dictionaries
```

### 7. Honest Notes on Scope

This is a complete, working reference implementation, not a prototype with placeholder UI — every engine listed above runs real logic against real data, and the whole stack was built and test-run end to end (telemetry flowing through the real pipeline, a simulated equipment failure cascading through the real dependency graph, the Command Safety Engine genuinely rejecting an unauthorized actuator command and accepting it after a permission grant, and so on).

Two things are intentionally left as extension points rather than finished for you, because they depend on infrastructure or hardware only you have:

- **External protocol adapters** (OPC-UA, Modbus/TCP, MQTT, Kafka, NATS, gRPC) are defined as clean Go interfaces (`telemetry.Adapter`, `actuator.Adapter`) with a fully working simulated/HTTP reference implementation. Wiring a real PLC or SCADA system in means implementing that one interface against your equipment.
- **Horizontal/distributed deployment**: the reference build runs as a single `riftd` process (with a Go worker pool per subsystem) plus an in-memory graph store with JSON-file persistence. Swapping in PostgreSQL/TimescaleDB and Kafka/NATS behind the same `registry`/`telemetry` interfaces is the documented path to a multi-node deployment; the storage boundary was designed for exactly that swap.

---

## 🇮🇷 فارسی

### ۱. ریفت چیست؟

ریفت یک محیط واقعی — کف یک کارخانه، یک برج اداری، یک دیتاسنتر، یک انبار، یا حتی یک منطقه شهری — را به یک **دوقلوی دیجیتال زنده** تبدیل می‌کند: مدلی ساختاریافته از تمام موجودیت‌ها، سنسورها، عملگرها و ارتباطات آن محیط، که همواره با داده‌های واقعی هم‌گام است، می‌تواند آینده خود را شبیه‌سازی کند، و پیش از هر تغییری در دنیای واقعی به سؤال «اگر… چه می‌شود؟» پاسخ دهد.

این پروژه از دو بخش تشکیل شده است:

| بخش | فناوری | مسئولیت |
|---|---|---|
| `backend/` | Go (فقط کتابخانه استاندارد) | هسته دوقلوی دیجیتال، گراف موجودیت/ارتباط، دریافت تله‌متری، موتور وضعیت، موتور رویداد، موتور قوانین، موتور شبیه‌سازی و فرضی، تحلیل‌ها، امنیت، API مبتنی بر REST و SSE، و رابط خط‌فرمان |
| `frontend/` | Next.js 14 · React 18 · TypeScript · Tailwind CSS | برج کنترل: داشبوردها، نمایشگر سه‌بعدی و دوبعدی دوقلو، کاوشگر تله‌متری، سازنده قوانین و سناریو، پیکربندی ساعات کاری، هشدارها، ردیابی رویدادها، مدیریت دسترسی‌ها |

### ۲. نمای کلی ویژگی‌ها

**هسته دوقلوی دیجیتال**
- مدل جهانی موجودیت (نوع، ویژگی‌ها، وضعیت، موقعیت/جهت/ابعاد، متادیتا، مالک، والد/فرزندان، سلامت، وضعیت عملیاتی، نسخه، زمان‌ها)
- ارتباطات تایپ‌شده (`contains`، `connected_to`، `powered_by`، `depends_on`، `located_in`، `controlled_by`، `communicates_with`، `feeds`، `produces`، `consumes`، `transports`، `monitors`) همراه با ضریب اطمینان و تاریخچه
- رجیستری دوقلو با قابلیت ایجاد، کلون، آرشیو، عکس‌گیری از وضعیت، برون‌بری و درون‌ریزی
- قالب‌های آماده: **کارخانه، ساختمان هوشمند، دیتاسنتر، انبار، شهر هوشمند** — همگی کاملاً قابل شخصی‌سازی

**تله‌متری و وضعیت**
- لایه انتزاع سنسور (دما، رطوبت، فشار، ارتعاش، انرژی، ولتاژ، جریان، کیفیت هوا، نور، حرکت، سرعت، موقعیت مکانی، CPU، حافظه، دیسک، شبکه و انواع سفارشی)
- خط‌لوله دریافت داده مبتنی بر گوروتین با مدیریت فشار برگشتی، بررسی کیفیت داده و تشخیص قطعی/آفلاین‌شدن سنسور
- موتور وضعیت چهارگانه: فعلی / قبلی / مورد‌انتظار / شبیه‌سازی‌شده
- ذخیره‌سازی سری‌زمانی با پنجره‌بندی و نمونه‌کاهی سطلی

**اتوماسیون و هوشمندی**
- موتور قوانین (محرک ← شرایط ← اقدامات، اولویت، فاصله زمانی، محافظ ایمنی)
- تشخیص ناهنجاری آماری (خط‌مبنای خودتنظیم برای هر سنسور)
- امتیازدهی سلامت / ریسک / بحرانی‌بودن برای هر موجودیت و کل دوقلو
- موتور گراف وابستگی (تحلیل اثر، کوتاه‌ترین مسیر، قابلیت دسترسی، عمق وابستگی)
- تحلیل علت ریشه‌ای (همبستگی رویدادها در سراسر گراف وابستگی)
- موتور هشداردهی (شدت، حذف تکرار، فاصله زمانی، تشدید، تأیید، اعلان از طریق webhook)

**شبیه‌سازی**
- موتور زمان (توقف / ادامه / گام / پرش / تغییر سرعت، مستقل از زمان واقعی)
- موتور سناریوی فرضی — هر شبیه‌سازی روی یک شاخه ایزوله در حافظه اجرا می‌شود و **هرگز وضعیت واقعی دوقلو را تغییر نمی‌دهد**
- تزریق خطا (خرابی سنسور، خرابی شبکه، قطعی برق، خرابی تجهیزات، تأخیر ارتباطی، افت بسته، اتمام منابع)
- موتور مونت‌کارلو و موتور پیمایش پارامتر
- موتور بهینه‌سازی چندهدفه (ترکیب هزینه، ریسک، توان عملیاتی، انرژی)

**امنیت و چندمستأجری**
- احراز هویت مبتنی بر نشست (توکن‌های امضاشده با HMAC) و ذخیره‌سازی گذرواژه با نمک و کشش رمزنگاری
- مدل دسترسی دوقلو به‌سبک RBAC و ABAC (در سطح نقش یا کاربر، با امکان محدودکردن به یک موجودیت خاص)
- موتور ایمنی فرمان پیش از اجرای هر دستور به عملگر: فهرست مجاز، بررسی دسترسی، بررسی محدوده مقدار، قفل‌های تداخلی، محدودیت نرخ، و مرحله تأیید صریح برای دستورات حساس
- ردیابی کامل رویدادها که روی دیسک ذخیره می‌شود

**برج کنترل (رابط کاربری)**
- داشبورد برج کنترل با شاخص‌های کلیدی زنده و توان عملیاتی خط‌لوله دریافت داده
- درخت موجودیت‌ها + بازرس، همراه با تحلیل اثر و جست‌وجوی علت ریشه‌ای با یک کلیک
- نمایشگر سه‌بعدی تعاملی دوقلو (Three.js / React Three Fiber) و یک نقشه عملیاتی دوبعدی بدون وابستگی بیرونی
- نمودارهای تله‌متری بلادرنگ از طریق Server-Sent Events (بدون نیاز به وابستگی انتقالی اضافه)
- سازنده قوانین، سازنده سناریوی فرضی همراه با نتایج مونت‌کارلو، ردیابی رویدادها، و مدیریت دسترسی‌ها
- **ساعات کاری قابل تنظیم توسط کاربر**: شما زمان باز و بسته‌شدن هر روز را خودتان وارد می‌کنید و ریفت وضعیت باز/بسته زنده به‌همراه شمارش معکوس تا تغییر بعدی را محاسبه می‌کند
- پنج پوسته — **ویندوز (پیش‌فرض)، روشن، تاریک، قرمز، آبی** — و بومی‌سازی کامل به **انگلیسی / فارسی / چینی**، با چیدمان خودکار راست‌به‌چپ برای فارسی

### ۳. پیش‌نیازها

- **Go نسخه ۱٫۲۲ یا جدیدتر** (بک‌اند هیچ وابستگی بیرونی ندارد — فقط کتابخانه استاندارد)
- **Node.js نسخه ۱۸٫۱۸ یا جدیدتر** و **npm** (فرانت‌اند)

### ۴. نصب و اجرا

**گام ۱ — اجرای بک‌اند**

```bash
cd backend
go run ./cmd/riftd
```

سرور به‌طور پیش‌فرض روی `http://localhost:8080` گوش می‌دهد و پوشه‌ی `data/` را برای ذخیره وضعیت و لاگ رویدادها می‌سازد. در اولین اجرا، یک دوقلوی نمایشی برای هر قالب آماده می‌سازد و یک حساب مدیر پیش‌فرض ایجاد می‌کند:

```
نام‌کاربری: admin
گذرواژه: admin123
```

در هر محیط اشتراکی یا عملیاتی، این گذرواژه را بلافاصله تغییر دهید.

پرچم‌ها/متغیرهای محیطی اختیاری:

```bash
go run ./cmd/riftd -addr :8080 -data ./data
# یا
RIFT_ADDR=:8080 RIFT_DATA_DIR=./data go run ./cmd/riftd
```

برای ساخت یک فایل اجرایی مستقل:

```bash
go build -o riftd ./cmd/riftd
./riftd
```

**گام ۲ — اجرای برج کنترل (فرانت‌اند)**

در یک ترمینال دوم:

```bash
cd frontend
npm install
cp .env.example .env.local   # به‌طور پیش‌فرض رابط کاربری را به http://localhost:8080 متصل می‌کند
npm run dev
```

آدرس `http://localhost:3000` را در مرورگر باز کنید، با حساب مدیر بالا وارد شوید، و پنج دوقلوی آماده (کارخانه، ساختمان هوشمند، دیتاسنتر، انبار، شهر هوشمند) را خواهید دید که در حال تولید تله‌متری شبیه‌سازی‌شده زنده هستند.

برای ساخت نسخه عملیاتی فرانت‌اند:

```bash
npm run build
npm run start
```

**گام ۳ — (اختیاری) استفاده از رابط خط‌فرمان**

```bash
cd backend
go build -o rift ./cmd/rift
export RIFT_SERVER=http://localhost:8080
./rift login --username=admin --password=admin123
export RIFT_TOKEN=<توکن چاپ‌شده در مرحله قبل را اینجا قرار دهید>
./rift twin list
./rift twin create --name="کارخانه من" --template=factory
./rift scenario create <twinId> --name="خرابی خنک‌کننده" --ticks=40 --fault=equipment_failure:<entityId>:0:1
./rift scenario run <scenarioId>
./rift benchmark <twinId> --count=100000
```

برای فهرست کامل دستورات (`twin`، `entity`، `sensor`، `ingest`، `simulate`، `scenario`، `snapshot`/`export`، `replay`، `analyze`، `optimize`، `benchmark`، `worker`) دستور `./rift help` را اجرا کنید.

### ۵. پیکربندی ساعات کاری

به تب **ساعات کاری** یک دوقلو بروید، برای برنامه یک برچسب و منطقه زمانی تعیین کنید، و زمان باز و بسته‌شدن هر روز هفته را خودتان وارد نمایید (یا یک روز را کاملاً تعطیل علامت بزنید). ریفت بلافاصله محاسبه می‌کند و به‌صورت زنده نگه می‌دارد که:

- آیا دوقلو اکنون باز است یا خیر،
- تغییر بعدی چه خواهد بود (باز شدن یا بسته شدن)، و
- دقیقاً چه‌مقدار زمان تا آن تغییر باقی مانده است.

هیچ‌چیز در این بخش از پیش با فرض خاصی پر نشده — تمام ساعات از همان چیزی می‌آید که شما وارد می‌کنید.

### ۶. ساختار پروژه

```
RIFT-Digital-Twin-Engine/
├── backend/
│   ├── cmd/riftd/           # نقطه ورود سرور HTTP و SSE
│   ├── cmd/rift/            # کلاینت خط‌فرمان
│   ├── internal/model/      # مدل جهانی دوقلو (موجودیت، سنسور، رویداد، قانون...)
│   ├── internal/registry/   # رجیستری دوقلو (ذخیره‌سازی گراف + پایداری)
│   ├── internal/telemetry/  # خط‌لوله دریافت، آداپترها، ذخیره‌سازی سری‌زمانی
│   ├── internal/stateengine/ internal/events/ internal/rules/ internal/anomaly/
│   ├── internal/health/ internal/graph/ internal/rootcause/
│   ├── internal/simulation/ internal/optimization/ internal/actuator/
│   ├── internal/alerting/ internal/auth/ internal/audit/
│   ├── internal/operatinghours/ internal/synthetic/ internal/exportimport/
│   └── internal/api/        # هندلرهای REST و SSE که همه موتورها را به‌هم متصل می‌کنند
└── frontend/
    └── src/
        ├── app/              # صفحات Next.js App Router
        ├── components/       # داشبورد، نمایشگرهای سه‌بعدی/دوبعدی، پنل‌ها، Providerها
        └── lib/              # کلاینت API، تایپ‌ها، دیکشنری‌های چندزبانه
```

### ۷. یادداشت صادقانه درباره محدوده پروژه

این یک پیاده‌سازی مرجع کامل و واقعاً کارکننده است، نه یک نمونه اولیه با رابط کاربری ساختگی — تمام موتورهای فهرست‌شده در بالا منطق واقعی روی داده واقعی اجرا می‌کنند، و کل پشته به‌صورت سرتاسری ساخته و آزمایش شده است (جریان واقعی تله‌متری در خط‌لوله، خرابی شبیه‌سازی‌شده تجهیزات که در گراف وابستگی واقعی پخش می‌شود، رد واقعی یک فرمان غیرمجاز توسط موتور ایمنی فرمان و پذیرش آن پس از اعطای دسترسی، و مواردی از این دست).

دو مورد عمداً به‌عنوان نقطه توسعه باقی مانده‌اند نه چیزی که از قبل برای شما تمام شده باشد، چون به زیرساخت یا سخت‌افزاری وابسته‌اند که فقط خود شما در اختیار دارید:

- **آداپترهای پروتکل بیرونی** (OPC-UA، Modbus/TCP، MQTT، Kafka، NATS، gRPC) به‌صورت اینترفیس‌های تمیز Go (`telemetry.Adapter`، `actuator.Adapter`) تعریف شده‌اند، همراه با یک پیاده‌سازی مرجع شبیه‌سازی‌شده/HTTP کاملاً کارکننده. اتصال یک PLC یا سیستم SCADA واقعی یعنی پیاده‌سازی همان یک اینترفیس برای تجهیزات شما.
- **استقرار توزیع‌شده/افقی**: نسخه مرجع به‌صورت یک فرآیند تکی `riftd` (با یک استخر کارگر Go برای هر زیرسیستم) به‌همراه یک ذخیره‌ساز گراف در حافظه با پایداری فایل JSON اجرا می‌شود. جایگزینی PostgreSQL/TimescaleDB و Kafka/NATS پشت همین اینترفیس‌های `registry`/`telemetry`، مسیر مستندشده برای استقرار چندگرهی است؛ مرز ذخیره‌سازی دقیقاً برای همین جایگزینی طراحی شده است.

---

## 🇨🇳 中文

### 一、RIFT 是什么？

RIFT 将现实世界中的某个环境——工厂车间、办公大楼、数据中心、仓库，乃至整座城市片区——转变为一个**实时数字孪生体**：一个由该环境中所有实体、传感器、执行器与相互关系构成的结构化模型，持续与真实遥测数据保持同步，能够模拟自身的未来状态，并在现实世界发生任何变化之前回答"如果……会怎样"这一问题。

本项目由两部分组成：

| 组件 | 技术栈 | 职责 |
|---|---|---|
| `backend/` | Go（仅标准库） | 数字孪生核心、实体/关系图、遥测接入、状态引擎、事件引擎、规则引擎、仿真与假设情景引擎、分析能力、安全机制、REST + SSE API、命令行工具 |
| `frontend/` | Next.js 14 · React 18 · TypeScript · Tailwind CSS | 控制中心：仪表盘、孪生体三维/二维查看器、遥测浏览器、规则与场景构建器、运营时间配置、告警、审计日志、权限管理 |

### 二、功能概览

**数字孪生核心**
- 通用实体模型（类型、属性、状态、位置/朝向/尺寸、元数据、所有者、父级/子级、健康度、运行状态、版本、时间戳）
- 带类型的关系（`contains`、`connected_to`、`powered_by`、`depends_on`、`located_in`、`controlled_by`、`communicates_with`、`feeds`、`produces`、`consumes`、`transports`、`monitors`），并附带置信度与历史记录
- 孪生体注册中心，支持创建、克隆、归档、快照、导出、导入
- 内置模板：**工厂、智能建筑、数据中心、仓库、智慧城市**——均可完全自定义

**遥测与状态**
- 传感器抽象层（温度、湿度、压力、振动、能耗、电压、电流、空气质量、光照、运动、速度、位置、CPU、内存、磁盘、网络及自定义类型）
- 基于 Goroutine 工作池的接入管道，具备背压处理、数据质量校验与断线/离线检测
- 四态状态引擎：当前 / 先前 / 预期 / 仿真
- 支持窗口查询与分桶降采样的时间序列存储

**自动化与智能分析**
- 规则引擎（触发条件 → 条件判断 → 执行动作，含优先级、冷却时间、安全防护）
- 统计异常检测（每个传感器自适应基线）
- 实体级与孪生体级的健康度／风险／关键性评分
- 依赖关系图引擎（影响分析、最短路径、可达性、依赖深度）
- 根因分析（跨依赖图的事件关联）
- 告警引擎（严重级别、去重、冷却、升级、确认、Webhook 通知）

**仿真能力**
- 时间引擎（暂停/继续/单步/跳转/倍速，独立于真实时钟运行）
- 假设情景引擎——每次仿真都运行在独立的内存分支上，**绝不影响真实孪生体状态**
- 故障注入（传感器故障、网络故障、电源故障、设备故障、通信延迟、丢包、资源耗尽）
- 蒙特卡洛引擎与参数扫描引擎
- 多目标优化引擎（综合权衡成本、风险、吞吐量、能耗等目标）

**安全与多租户**
- 基于会话的身份认证（HMAC 签名令牌）及加盐、多轮哈希的密码存储
- RBAC 与 ABAC 风格结合的孪生体权限模型（可按角色或用户授予，亦可限定到单个实体）
- 每条执行器指令前置的命令安全引擎：白名单校验、权限检查、数值范围检查、联锁保护、频率限制，以及针对敏感指令的显式二次确认
- 完整的审计日志，持久化写入磁盘

**控制中心（前端）**
- 控制塔仪表盘，展示实时关键指标与接入管道吞吐量
- 实体树与检查器，一键完成影响分析与根因查找
- 交互式**三维孪生体查看器**（Three.js / React Three Fiber）与无外部依赖的**二维运营地图**
- 基于 Server-Sent Events 的实时遥测图表（无需额外的传输层依赖）
- 规则构建器、带蒙特卡洛结果展示的假设情景构建器、审计日志、权限管理
- **用户自定义运营时间**：由您自行输入每天的开始与结束时间，RIFT 会实时计算当前是否处于营业状态，并倒计时至下一次状态变化
- 五种主题——**Windows（默认）、浅色、深色、红色、蓝色**——并完整支持**英文／波斯语／中文**三种语言，波斯语自动采用从右到左的排版

### 三、环境要求

- **Go 1.22 或更高版本**（后端零外部依赖，仅使用标准库）
- **Node.js 18.18 或更高版本**及 **npm**（前端）

### 四、安装与运行

**第一步 —— 启动后端**

```bash
cd backend
go run ./cmd/riftd
```

服务器默认监听 `http://localhost:8080`，并创建 `data/` 目录用于保存持久化状态与审计日志。首次启动时，系统会为每个内置模板生成一个演示孪生体，并创建一个默认管理员账户：

```
用户名：admin
密码：admin123
```

请在任何共享或生产环境中立即修改此密码。

可选的命令行参数／环境变量：

```bash
go run ./cmd/riftd -addr :8080 -data ./data
# 或者
RIFT_ADDR=:8080 RIFT_DATA_DIR=./data go run ./cmd/riftd
```

如需构建独立的可执行文件：

```bash
go build -o riftd ./cmd/riftd
./riftd
```

**第二步 —— 启动控制中心（前端）**

打开第二个终端窗口：

```bash
cd frontend
npm install
cp .env.example .env.local   # 默认将前端指向 http://localhost:8080
npm run dev
```

在浏览器中打开 `http://localhost:3000`，使用上面的管理员账户登录，即可看到五个已生成实时仿真遥测数据、可直接浏览的孪生体（工厂、智能建筑、数据中心、仓库、智慧城市）。

如需构建前端的生产版本：

```bash
npm run build
npm run start
```

**第三步 ——（可选）使用命令行工具**

```bash
cd backend
go build -o rift ./cmd/rift
export RIFT_SERVER=http://localhost:8080
./rift login --username=admin --password=admin123
export RIFT_TOKEN=<粘贴上一步打印出的令牌>
./rift twin list
./rift twin create --name="我的工厂" --template=factory
./rift scenario create <twinId> --name="制冷系统故障" --ticks=40 --fault=equipment_failure:<entityId>:0:1
./rift scenario run <scenarioId>
./rift benchmark <twinId> --count=100000
```

运行 `./rift help` 可查看完整命令参考（`twin`、`entity`、`sensor`、`ingest`、`simulate`、`scenario`、`snapshot`/`export`、`replay`、`analyze`、`optimize`、`benchmark`、`worker`）。

### 五、配置运营时间

打开某个孪生体的**运营时间**标签页，为该时间表设置标签与时区，并自行输入每周每一天的开始与结束营业时间（也可将某天标记为全天休息）。RIFT 会立即计算并实时保持以下信息：

- 该孪生体当前是否处于营业状态；
- 下一次状态变化将是什么（开始营业还是结束营业）；
- 距离该变化确切还剩多长时间。

这里的所有时间均不带任何预设假设——完全来自您输入的内容。

### 六、项目结构

```
RIFT-Digital-Twin-Engine/
├── backend/
│   ├── cmd/riftd/           # HTTP + SSE 服务器入口
│   ├── cmd/rift/            # 命令行客户端
│   ├── internal/model/      # 通用孪生体模型（实体、传感器、事件、规则……）
│   ├── internal/registry/   # 孪生体注册中心（图存储 + 持久化）
│   ├── internal/telemetry/  # 接入管道、适配器、时间序列存储
│   ├── internal/stateengine/ internal/events/ internal/rules/ internal/anomaly/
│   ├── internal/health/ internal/graph/ internal/rootcause/
│   ├── internal/simulation/ internal/optimization/ internal/actuator/
│   ├── internal/alerting/ internal/auth/ internal/audit/
│   ├── internal/operatinghours/ internal/synthetic/ internal/exportimport/
│   └── internal/api/        # 将所有引擎串联起来的 REST + SSE 处理器
└── frontend/
    └── src/
        ├── app/              # Next.js App Router 页面
        ├── components/       # 仪表盘、三维/二维查看器、功能面板、上下文提供者
        └── lib/              # API 客户端、类型定义、多语言词典
```

### 七、关于项目范围的诚实说明

这是一个完整且真正可运行的参考实现，而不是带有占位界面的原型——上述列出的每一个引擎都在真实数据上执行真实逻辑，整套系统已经过端到端的搭建与实测（真实遥测数据流经真实管道、模拟的设备故障在真实依赖关系图中级联传播、命令安全引擎真实地拒绝了一次未授权的执行器指令、并在授予权限后予以放行，等等）。

有两项内容被有意保留为扩展点，而非替您完成，因为它们依赖于只有您自己才拥有的基础设施或硬件：

- **外部协议适配器**（OPC-UA、Modbus/TCP、MQTT、Kafka、NATS、gRPC）已定义为简洁的 Go 接口（`telemetry.Adapter`、`actuator.Adapter`），并附带一个完全可用的模拟／HTTP 参考实现。接入真实的 PLC 或 SCADA 系统，只需针对您的设备实现这同一个接口即可。
- **水平/分布式部署**：参考实现以单一 `riftd` 进程运行（每个子系统配有独立的 Go 工作池），并使用内存图存储加 JSON 文件持久化。在相同的 `registry`／`telemetry` 接口之下替换为 PostgreSQL/TimescaleDB 与 Kafka/NATS，是通向多节点部署的既定路径；存储边界正是为这一替换而设计的。
