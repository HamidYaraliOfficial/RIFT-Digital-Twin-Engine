import type { Dictionary } from "./en";

const zh: Dictionary = {
  meta: { title: "RIFT — 数字孪生引擎", tagline: "通用数字孪生操作平台" },
  nav: {
    dashboard: "控制塔",
    twins: "孪生体",
    settings: "设置",
    login: "登录",
    logout: "退出登录",
  },
  common: {
    create: "创建", save: "保存", cancel: "取消", delete: "删除", close: "关闭",
    run: "运行", refresh: "刷新", loading: "加载中…", search: "搜索",
    name: "名称", type: "类型", status: "状态", actions: "操作", details: "详情",
    yes: "是", no: "否", of: "/", none: "暂无可用项", back: "返回",
  },
  status: {
    operational: "运行正常", warning: "警告", degraded: "性能下降",
    maintenance: "维护中", failed: "故障", recovered: "已恢复", unknown: "未知",
  },
  severity: { info: "信息", warning: "警告", critical: "严重" },
  dashboard: {
    title: "数字孪生控制塔",
    subtitle: "实时查看 RIFT 建模的所有环境状态。",
    createTwin: "新建孪生体",
    templates: { factory: "工厂", building: "智能建筑", data_center: "数据中心", warehouse: "仓库", smart_city: "智慧城市", custom: "自定义" },
    openTwin: "打开工作区",
    entities: "实体", sensors: "传感器", health: "整体健康度", alerts: "未处理告警",
    createDialogTitle: "创建新的数字孪生体",
    twinName: "孪生体名称", startFromTemplate: "从模板开始（可选）",
    pipeline: "数据接入管道", received: "已接收", processed: "已处理", dropped: "已丢弃",
  },
  workspace: {
    tabs: {
      overview: "概览", entities: "实体", view3d: "三维视图", map2d: "二维地图",
      telemetry: "遥测数据", rules: "规则", scenarios: "场景", hours: "运营时间",
      audit: "审计日志", permissions: "权限管理",
    },
    healthOverview: "健康度概览", recentAlerts: "最近告警", recentEvents: "最近事件",
    clock: "仿真时钟", pause: "暂停", resume: "继续", speed: "速度", step: "步进 +1分钟",
  },
  entity: {
    tree: "实体树", inspector: "检查器", noSelection: "请选择一个实体进行查看。",
    properties: "属性", sensors: "传感器", relationships: "关系", position: "位置",
    createEntity: "新建实体", parent: "父实体（可选）", newEntityDefaultParent: "无 — 顶层",
    impact: "影响分析", runImpact: "分析下游影响", dependents: "依赖此实体的实体",
    rootcause: "可能根因", runRootCause: "查找可能根因", confidence: "置信度",
    sendCommand: "发送执行器指令", commandType: "指令", commandValue: "数值（如适用）",
    commandReason: "原因", commandSubmit: "提交安全审核",
  },
  telemetry: {
    title: "实时遥测浏览器", selectSensor: "选择传感器",
    live: "实时", window: "采样窗口", noData: "该传感器暂无遥测数据。",
  },
  rules: {
    title: "规则引擎", newRule: "新建规则", trigger: "触发条件（传感器类型）", field: "字段",
    operator: "运算符", value: "数值", action: "匹配时的动作", cooldown: "冷却时间（毫秒）",
    enabled: "已启用", empty: "该孪生体尚未定义任何规则。",
  },
  scenarios: {
    title: "假设情景引擎", newScenario: "新建场景", scenarioName: "场景名称",
    durationTicks: "持续时长（tick）", addFault: "添加故障注入", faultType: "故障类型",
    targetEntity: "目标实体", atTick: "触发时刻（tick）", magnitude: "强度（0–1）",
    runScenario: "运行仿真", results: "结果", affected: "受影响实体",
    summary: "摘要", empty: "尚无场景 — 创建一个以运行假设仿真。",
    montecarlo: "蒙特卡洛", optimize: "优化", trials: "试验次数",
  },
  hours: {
    title: "运营时间", subtitle: "自行输入时间表 — RIFT 会计算实时的开/关状态以及距下次变化的倒计时。",
    label: "时间表标签", timezone: "时区（IANA，例如 Asia/Baku）",
    openNow: "当前营业", closedNow: "当前休息", nextChange: "下次变化", in: "还有",
    closedAllDay: "全天休息", open: "开始", close: "结束", save: "保存时间表",
    days: ["周日", "周一", "周二", "周三", "周四", "周五", "周六"],
  },
  alerts: {
    title: "告警", acknowledge: "确认", resolve: "解决", empty: "没有未处理的告警，一切正常。",
    occurrences: "次发生",
  },
  audit: { title: "审计日志", actor: "操作者", action: "动作", target: "目标", when: "时间" },
  permissions: {
    title: "孪生体权限模型", grant: "授予权限", user: "用户ID（可选）", role: "角色（可选）",
    entityScope: "限定到单个实体（可选）", scopes: "权限范围", scopeRead: "读取", scopeWrite: "写入",
    scopeScenario: "运行场景", scopeActuator: "控制执行器", scopeAdmin: "完全管理",
  },
  auth: {
    title: "登录 RIFT", username: "用户名", password: "密码", signIn: "登录",
    defaultHint: "默认安装账户：admin / admin123", error: "用户名或密码不正确。",
  },
  settings: {
    title: "设置", appearance: "外观", theme: "主题", language: "语言",
    themes: { windows: "Windows（默认）", light: "浅色", dark: "深色", red: "红色", blue: "蓝色" },
    languages: { en: "英语 (English)", fa: "波斯语 (فارسی)", zh: "中文" },
  },
};

export default zh;
