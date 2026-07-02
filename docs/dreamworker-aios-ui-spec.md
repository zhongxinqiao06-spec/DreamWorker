# DreamWorker AI OS 2.0 UI 规范

本规范从用户提供的 5 张参考图沉淀而来，目标是统一 DreamWorker 桌面前端的“银白玻璃 AI OS”视觉语言，并作为本轮前端改造的 token 来源。

## 1. 品牌色

| Token            | 用途                         | Hex       |
| ---------------- | ---------------------------- | --------- |
| `--brand-purple` | 主按钮、选中态、品牌渐变起点 | `#7C3AED` |
| `--brand-blue`   | Chat、信息态、品牌渐变中段   | `#2563EB` |
| `--brand-teal`   | 探索、成功运行、高亮数据     | `#14B8A6` |
| `--brand-lilac`  | 柔和紫雾、浅色背景装饰       | `#EDE7FF` |

## 2. 中性色

| Token              | 用途       | Hex       |
| ------------------ | ---------- | --------- |
| `--white`          | 主背景高光 | `#FFFFFF` |
| `--silver`         | 页面底色   | `#F8FAFC` |
| `--mist`           | 分区底色   | `#F1F5F9` |
| `--border`         | 玻璃边框   | `#E2E8F0` |
| `--text-main`      | 主文本     | `#0F172A` |
| `--text-secondary` | 次文本     | `#64748B` |
| `--text-tertiary`  | 辅助说明   | `#94A3B8` |

## 3. 模块色

资源中心 `#8B5CF6`，Chat `#2563EB`，探索 `#14B8A6`，产品 `#F59E0B`，开发 `#2F80ED`，销售 `#EC4899`，诊断 `#10B981`。

## 4. 语义色

成功 `#10B981`，警告 `#F59E0B`，错误 `#EF4444`，信息 `#2563EB`。语义色默认使用 10%-14% 透明底，边框透明度 22%-36%，正文使用原色或 8% 加深色。

## 5. 组件规范

- 页面背景：银白渐变底，右上角玻璃环装饰，局部紫蓝/青绿柔光。
- 容器：玻璃面板 `rgba(255,255,255,.72-.88)`，1px 浅边框，`backdrop-filter: blur(18px)`。
- 卡片：8px 圆角，轻阴影，边框使用 `rgba(148,163,184,.18-.28)`。
- 主按钮：紫蓝渐变，白字，内阴影高光，hover 轻微上浮。
- 次按钮/图标按钮：白色玻璃底，浅边框，hover 切到淡紫底。
- 输入框：白色半透明底，聚焦时紫蓝边框与浅紫外发光。
- Tabs/Chips：胶囊形，小字号，选中态使用淡紫底 + 主紫文本。
- 状态 Badge：成功/警告/错误/信息使用语义色浅底。
- 字体：系统无衬线优先，中文界面保持清晰紧凑，不使用负字距。

## 6. 本地切图

静态资产已放入 `apps/desktop/renderer/public/aios/`：

- `brand-lockup.png`：品牌完整标识。
- `brand-mark.png`：品牌图标。
- `glass-orbit-hero.png`：顶部/开屏玻璃环。
- `glass-orbit-corner.png`：角落玻璃装饰。
- `resource-orbit-banner.png`：资源页横幅装饰。
- `empty-glass-prism.png`：空态玻璃插画。
- `dreamworker-aios-style-board.png`：色卡与规范参考图。
- `dreamworker-aios-component-board.png`：组件资产参考图。
