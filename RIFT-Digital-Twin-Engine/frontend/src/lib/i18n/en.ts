const en = {
  meta: { title: "RIFT — Digital Twin Engine", tagline: "Universal Digital Twin Operating Platform" },
  nav: {
    dashboard: "Control Tower",
    twins: "Twins",
    settings: "Settings",
    login: "Sign in",
    logout: "Sign out",
  },
  common: {
    create: "Create", save: "Save", cancel: "Cancel", delete: "Delete", close: "Close",
    run: "Run", refresh: "Refresh", loading: "Loading…", search: "Search",
    name: "Name", type: "Type", status: "Status", actions: "Actions", details: "Details",
    yes: "Yes", no: "No", of: "of", none: "None available", back: "Back",
  },
  status: {
    operational: "Operational", warning: "Warning", degraded: "Degraded",
    maintenance: "Maintenance", failed: "Failed", recovered: "Recovered", unknown: "Unknown",
  },
  severity: { info: "Info", warning: "Warning", critical: "Critical" },
  dashboard: {
    title: "Digital Twin Control Tower",
    subtitle: "Live status across every environment RIFT is modeling.",
    createTwin: "New Twin",
    templates: { factory: "Factory", building: "Smart Building", data_center: "Data Center", warehouse: "Warehouse", smart_city: "Smart City", custom: "Custom" },
    openTwin: "Open workspace",
    entities: "Entities", sensors: "Sensors", health: "Overall Health", alerts: "Open Alerts",
    createDialogTitle: "Create a new Digital Twin",
    twinName: "Twin name", startFromTemplate: "Start from a template (optional)",
    pipeline: "Ingestion Pipeline", received: "Received", processed: "Processed", dropped: "Dropped",
  },
  workspace: {
    tabs: {
      overview: "Overview", entities: "Entities", view3d: "3D View", map2d: "2D Map",
      telemetry: "Telemetry", rules: "Rules", scenarios: "Scenarios", hours: "Operating Hours",
      audit: "Audit Trail", permissions: "Permissions",
    },
    healthOverview: "Health Overview", recentAlerts: "Recent Alerts", recentEvents: "Recent Events",
    clock: "Simulation Clock", pause: "Pause", resume: "Resume", speed: "Speed", step: "Step +1m",
  },
  entity: {
    tree: "Entity Tree", inspector: "Inspector", noSelection: "Select an entity to inspect it.",
    properties: "Properties", sensors: "Sensors", relationships: "Relationships", position: "Position",
    createEntity: "New Entity", parent: "Parent (optional)", newEntityDefaultParent: "None — top level",
    impact: "Impact Analysis", runImpact: "Analyze downstream impact", dependents: "Entities that depend on this one",
    rootcause: "Root Cause Candidates", runRootCause: "Find likely root causes", confidence: "Confidence",
    sendCommand: "Send Actuator Command", commandType: "Command", commandValue: "Value (if applicable)",
    commandReason: "Reason", commandSubmit: "Submit for safety review",
  },
  telemetry: {
    title: "Real-Time Telemetry Explorer", selectSensor: "Select a sensor",
    live: "Live", window: "Sample window", noData: "No telemetry yet for this sensor.",
  },
  rules: {
    title: "Rules Engine", newRule: "New Rule", trigger: "Trigger (sensor type)", field: "Field",
    operator: "Operator", value: "Value", action: "Action on match", cooldown: "Cooldown (ms)",
    enabled: "Enabled", empty: "No rules defined yet for this twin.",
  },
  scenarios: {
    title: "What-If Scenario Engine", newScenario: "New Scenario", scenarioName: "Scenario name",
    durationTicks: "Duration (ticks)", addFault: "Add fault injection", faultType: "Fault type",
    targetEntity: "Target entity", atTick: "At tick", magnitude: "Magnitude (0–1)",
    runScenario: "Run simulation", results: "Results", affected: "Affected entities",
    summary: "Summary", empty: "No scenarios yet — create one to run a What-If simulation.",
    montecarlo: "Monte Carlo", optimize: "Optimization", trials: "Trials",
  },
  hours: {
    title: "Operating Hours", subtitle: "Enter the schedule yourself — RIFT computes live open/closed status and the countdown to the next change.",
    label: "Schedule label", timezone: "Time zone (IANA, e.g. Asia/Baku)",
    openNow: "Open now", closedNow: "Closed now", nextChange: "Next change", in: "in",
    closedAllDay: "Closed all day", open: "Open", close: "Close", save: "Save schedule",
    days: ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"],
  },
  alerts: {
    title: "Alerts", acknowledge: "Acknowledge", resolve: "Resolve", empty: "No open alerts. Everything looks healthy.",
    occurrences: "occurrences",
  },
  audit: { title: "Audit Trail", actor: "Actor", action: "Action", target: "Target", when: "When" },
  permissions: {
    title: "Twin Permission Model", grant: "Grant access", user: "User ID (optional)", role: "Role (optional)",
    entityScope: "Limit to one entity (optional)", scopes: "Scopes", scopeRead: "Read", scopeWrite: "Write",
    scopeScenario: "Run scenarios", scopeActuator: "Control actuators", scopeAdmin: "Full admin",
  },
  auth: {
    title: "Sign in to RIFT", username: "Username", password: "Password", signIn: "Sign in",
    defaultHint: "Default install: admin / admin123", error: "Invalid username or password.",
  },
  settings: {
    title: "Settings", appearance: "Appearance", theme: "Theme", language: "Language",
    themes: { windows: "Windows (default)", light: "Light", dark: "Dark", red: "Red", blue: "Blue" },
    languages: { en: "English", fa: "Persian (فارسی)", zh: "Chinese (中文)" },
  },
};

export default en;
export type Dictionary = typeof en;
