import type { Dictionary } from "./en";

const fa: Dictionary = {
  meta: { title: "ریفت — موتور دوقلوی دیجیتال", tagline: "پلتفرم عملیاتی جهانی دوقلوی دیجیتال" },
  nav: {
    dashboard: "برج کنترل",
    twins: "دوقلوها",
    settings: "تنظیمات",
    login: "ورود",
    logout: "خروج",
  },
  common: {
    create: "ایجاد", save: "ذخیره", cancel: "انصراف", delete: "حذف", close: "بستن",
    run: "اجرا", refresh: "به‌روزرسانی", loading: "در حال بارگذاری…", search: "جستجو",
    name: "نام", type: "نوع", status: "وضعیت", actions: "عملیات", details: "جزئیات",
    yes: "بله", no: "خیر", of: "از", none: "چیزی موجود نیست", back: "بازگشت",
  },
  status: {
    operational: "عملیاتی", warning: "هشدار", degraded: "کاهش‌کارایی",
    maintenance: "تعمیرات", failed: "خرابی", recovered: "بازیابی‌شده", unknown: "نامشخص",
  },
  severity: { info: "اطلاعاتی", warning: "هشدار", critical: "بحرانی" },
  dashboard: {
    title: "برج کنترل دوقلوی دیجیتال",
    subtitle: "وضعیت زنده تمام محیط‌هایی که ریفت مدل‌سازی می‌کند.",
    createTwin: "دوقلوی جدید",
    templates: { factory: "کارخانه", building: "ساختمان هوشمند", data_center: "دیتاسنتر", warehouse: "انبار", smart_city: "شهر هوشمند", custom: "سفارشی" },
    openTwin: "باز کردن فضای کاری",
    entities: "موجودیت‌ها", sensors: "سنسورها", health: "سلامت کلی", alerts: "هشدارهای باز",
    createDialogTitle: "ایجاد یک دوقلوی دیجیتال جدید",
    twinName: "نام دوقلو", startFromTemplate: "شروع از یک قالب (اختیاری)",
    pipeline: "خط لوله دریافت داده", received: "دریافت‌شده", processed: "پردازش‌شده", dropped: "افت‌داده",
  },
  workspace: {
    tabs: {
      overview: "نمای‌کلی", entities: "موجودیت‌ها", view3d: "نمای سه‌بعدی", map2d: "نقشه دوبعدی",
      telemetry: "تله‌متری", rules: "قوانین", scenarios: "سناریوها", hours: "ساعات کاری",
      audit: "ردیابی رویدادها", permissions: "دسترسی‌ها",
    },
    healthOverview: "نمای‌کلی سلامت", recentAlerts: "هشدارهای اخیر", recentEvents: "رویدادهای اخیر",
    clock: "ساعت شبیه‌سازی", pause: "توقف", resume: "ادامه", speed: "سرعت", step: "گام +۱دقیقه",
  },
  entity: {
    tree: "درخت موجودیت‌ها", inspector: "بازرس", noSelection: "برای بازرسی، یک موجودیت انتخاب کنید.",
    properties: "ویژگی‌ها", sensors: "سنسورها", relationships: "ارتباطات", position: "موقعیت",
    createEntity: "موجودیت جدید", parent: "والد (اختیاری)", newEntityDefaultParent: "هیچ‌کدام — سطح بالا",
    impact: "تحلیل اثر", runImpact: "تحلیل اثر پایین‌دستی", dependents: "موجودیت‌هایی که به این وابسته‌اند",
    rootcause: "علل ریشه‌ای محتمل", runRootCause: "یافتن علل ریشه‌ای محتمل", confidence: "اطمینان",
    sendCommand: "ارسال فرمان به عملگر", commandType: "فرمان", commandValue: "مقدار (در صورت نیاز)",
    commandReason: "دلیل", commandSubmit: "ارسال برای بررسی ایمنی",
  },
  telemetry: {
    title: "کاوشگر تله‌متری بلادرنگ", selectSensor: "انتخاب سنسور",
    live: "زنده", window: "بازه نمونه‌برداری", noData: "هنوز داده‌ای برای این سنسور ثبت نشده است.",
  },
  rules: {
    title: "موتور قوانین", newRule: "قانون جدید", trigger: "محرک (نوع سنسور)", field: "فیلد",
    operator: "عملگر", value: "مقدار", action: "اقدام هنگام تطابق", cooldown: "فاصله زمانی (میلی‌ثانیه)",
    enabled: "فعال", empty: "هنوز قانونی برای این دوقلو تعریف نشده است.",
  },
  scenarios: {
    title: "موتور سناریوی فرضی", newScenario: "سناریوی جدید", scenarioName: "نام سناریو",
    durationTicks: "مدت (تیک)", addFault: "افزودن تزریق خطا", faultType: "نوع خطا",
    targetEntity: "موجودیت هدف", atTick: "در تیک", magnitude: "شدت (۰ تا ۱)",
    runScenario: "اجرای شبیه‌سازی", results: "نتایج", affected: "موجودیت‌های تحت‌تأثیر",
    summary: "خلاصه", empty: "هنوز سناریویی وجود ندارد — یکی بسازید تا شبیه‌سازی فرضی اجرا شود.",
    montecarlo: "مونت‌کارلو", optimize: "بهینه‌سازی", trials: "تعداد آزمایش",
  },
  hours: {
    title: "ساعات کاری", subtitle: "برنامه را خودتان وارد کنید — ریفت وضعیت باز/بسته زنده و شمارش معکوس تا تغییر بعدی را محاسبه می‌کند.",
    label: "برچسب برنامه", timezone: "منطقه زمانی (IANA، مثل Asia/Baku)",
    openNow: "اکنون باز است", closedNow: "اکنون بسته است", nextChange: "تغییر بعدی", in: "در",
    closedAllDay: "تمام روز بسته", open: "بازشدن", close: "بسته‌شدن", save: "ذخیره برنامه",
    days: ["یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه", "شنبه"],
  },
  alerts: {
    title: "هشدارها", acknowledge: "تأیید", resolve: "رفع", empty: "هیچ هشدار بازی وجود ندارد. همه‌چیز سالم است.",
    occurrences: "تکرار",
  },
  audit: { title: "ردیابی رویدادها", actor: "عامل", action: "اقدام", target: "هدف", when: "زمان" },
  permissions: {
    title: "مدل دسترسی دوقلو", grant: "اعطای دسترسی", user: "شناسه کاربر (اختیاری)", role: "نقش (اختیاری)",
    entityScope: "محدود به یک موجودیت (اختیاری)", scopes: "حوزه‌های دسترسی", scopeRead: "خواندن", scopeWrite: "نوشتن",
    scopeScenario: "اجرای سناریو", scopeActuator: "کنترل عملگرها", scopeAdmin: "مدیریت کامل",
  },
  auth: {
    title: "ورود به ریفت", username: "نام‌کاربری", password: "گذرواژه", signIn: "ورود",
    defaultHint: "نصب پیش‌فرض: admin / admin123", error: "نام‌کاربری یا گذرواژه نادرست است.",
  },
  settings: {
    title: "تنظیمات", appearance: "ظاهر", theme: "پوسته", language: "زبان",
    themes: { windows: "ویندوز (پیش‌فرض)", light: "روشن", dark: "تاریک", red: "قرمز", blue: "آبی" },
    languages: { en: "انگلیسی (English)", fa: "فارسی", zh: "چینی (中文)" },
  },
};

export default fa;
