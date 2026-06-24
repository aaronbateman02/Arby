export const venues = [
  { name: "Kalshi", cash: 42580.00, positions: 127350.00, portfolio: 169930.00 },
  { name: "Polymarket", cash: 18200.00, positions: 38400.00, portfolio: 56600.00 },
]

export const timePeriods = ["1d", "7d", "30d", "365d", "YTD"]

export const periodStats: Record<string, { openPositions: number; projectedProfit: number; roi: number; successfulFills: number; aborted: number }> = {
  "1d":   { openPositions: 12, projectedProfit: 340.00, roi: 2.1, successfulFills: 8, aborted: 1 },
  "7d":   { openPositions: 34, projectedProfit: 2180.00, roi: 4.3, successfulFills: 27, aborted: 5 },
  "30d":  { openPositions: 89, projectedProfit: 8450.00, roi: 6.8, successfulFills: 74, aborted: 18 },
  "365d": { openPositions: 156, projectedProfit: 42300.00, roi: 12.4, successfulFills: 412, aborted: 63 },
  "YTD":  { openPositions: 112, projectedProfit: 28900.00, roi: 9.1, successfulFills: 286, aborted: 41 },
}

export const bundleLegs = [
  { market: "BTC > $120K in June?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.31, actualCost: 0.32, fees: 0.42, status: "filled" as const },
  { market: "BTC > $120K in June?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.27, actualCost: 0.26, fees: 0.35, status: "filled" as const },
  { market: "Will DOW hit 45K by July?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.42, actualCost: 0.43, fees: 0.55, status: "filled" as const },
  { market: "Will DOW hit 45K by July?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.38, actualCost: 0.37, fees: 0.48, status: "filled" as const },
  { market: "Fed cuts rates in Q3?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.65, actualCost: 0.64, fees: 0.82, status: "filled" as const },
  { market: "Fed cuts rates in Q3?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.61, actualCost: 0.62, fees: 0.78, status: "filled" as const },
  { market: "S&P 500 > 5600 EOM?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.48, actualCost: 0.47, fees: 0.60, status: "filled" as const },
  { market: "S&P 500 > 5600 EOM?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.44, actualCost: 0.45, fees: 0.58, status: "filled" as const },
  { market: "ETH ETF approved by Aug?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.55, actualCost: 0.56, fees: 0.70, status: "filled" as const },
  { market: "ETH ETF approved by Aug?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.50, actualCost: 0.48, fees: 0.62, status: "filled" as const },
  { market: "Apple > $250 by Sept?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.37, actualCost: 0.36, fees: 0.48, status: "filled" as const },
  { market: "Apple > $250 by Sept?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.33, actualCost: 0.34, fees: 0.45, status: "filled" as const },
  { market: "US GDP growth > 2%?", venue: "Kalshi" as const, side: "Buy NO" as const, estimatedCost: 0.42, actualCost: 0.43, fees: 0.55, status: "filled" as const },
  { market: "US GDP growth > 2%?", venue: "Polymarket" as const, side: "Buy YES" as const, estimatedCost: 0.58, actualCost: 0.57, fees: 0.72, status: "filled" as const },
  { market: "Oil > $90 by Dec?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.35, actualCost: 0.36, fees: 0.47, status: "working" as const },
  { market: "Oil > $90 by Dec?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.31, actualCost: 0.30, fees: 0.40, status: "filled" as const },
  { market: "NFLX subscriber beat?", venue: "Kalshi" as const, side: "Buy YES" as const, estimatedCost: 0.62, actualCost: 0.63, fees: 0.78, status: "filled" as const },
  { market: "NFLX subscriber beat?", venue: "Polymarket" as const, side: "Buy NO" as const, estimatedCost: 0.40, actualCost: 0.39, fees: 0.52, status: "filled" as const },
  { market: "TSLA > $500 by Dec?", venue: "Kalshi" as const, side: "Buy NO" as const, estimatedCost: 0.55, actualCost: 0.54, fees: 0.70, status: "filled" as const },
  { market: "TSLA > $500 by Dec?", venue: "Polymarket" as const, side: "Buy YES" as const, estimatedCost: 0.45, actualCost: 0.46, fees: 0.60, status: "working" as const },
]

export const bundles = [
  {
    id: "bnd-001", name: "BTC June Cross", strategy: "Spread Capture v2", status: "active" as const,
    exposure: 1200, projectedRoi: 4.0, actualRoi: 3.8, totalFees: 3.74,
    openedAt: "2026-06-18 09:23:14 UTC", resolvesAt: "2026-06-30 12:00:00 UTC",
    pnl: 45.60, legs: bundleLegs.slice(0, 2),
  },
  {
    id: "bnd-002", name: "DOW July Spread", strategy: "Spread Capture v2", status: "active" as const,
    exposure: 850, projectedRoi: 4.0, actualRoi: 3.5, totalFees: 2.72,
    openedAt: "2026-06-17 14:05:33 UTC", resolvesAt: "2026-07-03 12:00:00 UTC",
    pnl: 29.75, legs: bundleLegs.slice(2, 4),
  },
  {
    id: "bnd-003", name: "Fed Q3 Rate Hedge", strategy: "Macro Arbitrage", status: "active" as const,
    exposure: 2000, projectedRoi: 3.5, actualRoi: 3.2, totalFees: 5.60,
    openedAt: "2026-06-15 11:42:01 UTC", resolvesAt: "2026-07-05 12:00:00 UTC",
    pnl: 64.00, legs: bundleLegs.slice(4, 6),
  },
  {
    id: "bnd-004", name: "S&P 500 EOM Play", strategy: "Spread Capture v2", status: "active" as const,
    exposure: 650, projectedRoi: 4.0, actualRoi: 3.6, totalFees: 2.36,
    openedAt: "2026-06-16 16:30:45 UTC", resolvesAt: "2026-07-10 12:00:00 UTC",
    pnl: 23.40, legs: bundleLegs.slice(6, 8),
  },
  {
    id: "bnd-005", name: "ETH ETF Approval", strategy: "Event Driven", status: "active" as const,
    exposure: 2100, projectedRoi: 5.0, actualRoi: 4.2, totalFees: 4.62,
    openedAt: "2026-06-14 08:15:22 UTC", resolvesAt: "2026-08-01 12:00:00 UTC",
    pnl: 88.20, legs: bundleLegs.slice(8, 10),
  },
  {
    id: "bnd-006", name: "Apple Sept Call", strategy: "Spread Capture v2", status: "active" as const,
    exposure: 400, projectedRoi: 4.0, actualRoi: 3.3, totalFees: 1.86,
    openedAt: "2026-06-19 10:44:18 UTC", resolvesAt: "2026-09-15 12:00:00 UTC",
    pnl: 13.20, legs: bundleLegs.slice(10, 12),
  },
  {
    id: "bnd-007", name: "GDP Growth Divergence", strategy: "Macro Arbitrage", status: "active" as const,
    exposure: 1500, projectedRoi: 3.0, actualRoi: 2.8, totalFees: 3.81,
    openedAt: "2026-06-12 07:20:55 UTC", resolvesAt: "2026-09-30 12:00:00 UTC",
    pnl: 42.00, legs: bundleLegs.slice(12, 14),
  },
  {
    id: "bnd-008", name: "Oil Dec Wager", strategy: "Commodity Arb", status: "active" as const,
    exposure: 900, projectedRoi: 4.0, actualRoi: 3.4, totalFees: 2.61,
    openedAt: "2026-06-20 13:12:09 UTC", resolvesAt: "2026-12-01 12:00:00 UTC",
    pnl: 30.60, legs: bundleLegs.slice(14, 16),
  },
  {
    id: "bnd-009", name: "NFLX Earnings Beat", strategy: "Event Driven", status: "active" as const,
    exposure: 1100, projectedRoi: 3.5, actualRoi: 3.1, totalFees: 3.25,
    openedAt: "2026-06-10 09:05:30 UTC", resolvesAt: "2026-10-15 12:00:00 UTC",
    pnl: 34.10, legs: bundleLegs.slice(16, 18),
  },
  {
    id: "bnd-010", name: "TSLA Year End", strategy: "Spread Capture v2", status: "active" as const,
    exposure: 1750, projectedRoi: 4.0, actualRoi: 3.7, totalFees: 4.55,
    openedAt: "2026-06-11 15:38:44 UTC", resolvesAt: "2026-12-15 12:00:00 UTC",
    pnl: 64.75, legs: bundleLegs.slice(18, 20),
  },
  {
    id: "bnd-011", name: "Target June Fill", strategy: "Spread Capture v2", status: "completed" as const,
    exposure: 600, projectedRoi: 3.8, actualRoi: 4.1, totalFees: 1.56,
    openedAt: "2026-06-01 10:00:00 UTC", resolvesAt: "2026-06-20 12:00:00 UTC",
    pnl: 24.60, legs: bundleLegs.slice(0, 2),
  },
  {
    id: "bnd-012", name: "DOW Early Close", strategy: "Spread Capture v2", status: "completed" as const,
    exposure: 500, projectedRoi: 3.5, actualRoi: 3.9, totalFees: 1.30,
    openedAt: "2026-05-28 08:22:15 UTC", resolvesAt: "2026-06-15 12:00:00 UTC",
    pnl: 19.50, legs: bundleLegs.slice(2, 4),
  },
  {
    id: "bnd-013", name: "ETH Pre-Approval", strategy: "Event Driven", status: "completed" as const,
    exposure: 1600, projectedRoi: 5.5, actualRoi: 6.2, totalFees: 4.16,
    openedAt: "2026-05-15 14:30:00 UTC", resolvesAt: "2026-06-10 12:00:00 UTC",
    pnl: 99.20, legs: bundleLegs.slice(8, 10),
  },
  {
    id: "bnd-014", name: "GDP Q2 Prediction", strategy: "Macro Arbitrage", status: "completed" as const,
    exposure: 1200, projectedRoi: 3.0, actualRoi: 3.3, totalFees: 3.12,
    openedAt: "2026-04-20 09:45:00 UTC", resolvesAt: "2026-05-30 12:00:00 UTC",
    pnl: 39.60, legs: bundleLegs.slice(12, 14),
  },
  {
    id: "bnd-015", name: "Oil Mid-Year", strategy: "Commodity Arb", status: "completed" as const,
    exposure: 750, projectedRoi: 4.0, actualRoi: 4.5, totalFees: 1.95,
    openedAt: "2026-05-01 11:00:00 UTC", resolvesAt: "2026-06-15 12:00:00 UTC",
    pnl: 33.75, legs: bundleLegs.slice(14, 16),
  },
].sort((a, b) => new Date(a.resolvesAt).getTime() - new Date(b.resolvesAt).getTime())

export const strategies = {
  best: { name: "Spread Capture v2", totalPnL: 18720, totalReturn: 14.2, winRate: 87, totalBundles: 64, successfulBundles: 56 },
  worst: { name: "Commodity Arb", totalPnL: 3240, totalReturn: 4.8, winRate: 62, totalBundles: 21, successfulBundles: 13 },
}

export const strategiesByPeriod: Record<string, typeof strategies> = {
  "1d":   { best: { name: "Event Driven", totalPnL: 180, totalReturn: 3.2, winRate: 100, totalBundles: 2, successfulBundles: 2 }, worst: { name: "Commodity Arb", totalPnL: -45, totalReturn: -1.1, winRate: 0, totalBundles: 1, successfulBundles: 0 } },
  "7d":   { best: { name: "Spread Capture v2", totalPnL: 1250, totalReturn: 5.8, winRate: 91, totalBundles: 11, successfulBundles: 10 }, worst: { name: "Macro Arbitrage", totalPnL: 210, totalReturn: 1.2, winRate: 67, totalBundles: 3, successfulBundles: 2 } },
  "30d":  { best: { name: "Spread Capture v2", totalPnL: 5200, totalReturn: 8.4, winRate: 89, totalBundles: 28, successfulBundles: 25 }, worst: { name: "Commodity Arb", totalPnL: 890, totalReturn: 3.1, winRate: 60, totalBundles: 10, successfulBundles: 6 } },
  "365d": { best: { name: "Event Driven", totalPnL: 18400, totalReturn: 16.5, winRate: 92, totalBundles: 38, successfulBundles: 35 }, worst: { name: "Commodity Arb", totalPnL: 4100, totalReturn: 5.2, winRate: 58, totalBundles: 24, successfulBundles: 14 } },
  "YTD":  { best: { name: "Spread Capture v2", totalPnL: 14200, totalReturn: 11.8, winRate: 88, totalBundles: 45, successfulBundles: 40 }, worst: { name: "Commodity Arb", totalPnL: 2800, totalReturn: 4.1, winRate: 61, totalBundles: 18, successfulBundles: 11 } },
}

export type Bundle = (typeof bundles)[number]

export interface StrategyConfig {
  minRoi: number
  maxPositionsPerBundle: number
  maxPositionDollars: number
  minSpread: number
  maxDailyExposure: number
  cooldownMinutes: number
  autoExecute: boolean
  notifyOnMatch: boolean
  maxDaysToResolution: number
  enabled: boolean
  paperMode: boolean
}

export interface MonitoredPair {
  id: string
  marketA: { title: string; venue: string; dir: string; price: number }
  marketB: { title: string; venue: string; dir: string; price: number }
  currentRoi: number
  currentSpread: number
  costPerShare: number
  expiresIn: string
  meetsCriteria: boolean
  executed: boolean
  estimatedProfit: number
}

export interface StrategyDetail {
  id: string
  name: string
  type: string
  description: string
  status: "active" | "paused" | "disabled"
  pausedReason?: string
  config: StrategyConfig
  stats: {
    totalBundles: number
    completedBundles: number
    abortedBundles: number
    totalPnL: number
    winRate: number
  }
  monitoredPairs: MonitoredPair[]
}

const basePairs: MonitoredPair[] = [
  { id: "pr-001", marketA: { title: "BTC > $120K in June?", venue: "Kalshi", dir: "BUY YES", price: 0.31 }, marketB: { title: "BTC > $120K in June?", venue: "Polymarket", dir: "BUY NO", price: 0.27 }, currentRoi: 4.0, currentSpread: 4.0, costPerShare: 0.58, expiresIn: "6d", meetsCriteria: true, executed: true, estimatedProfit: 40.00 },
  { id: "pr-002", marketA: { title: "Will DOW hit 45K by July?", venue: "Kalshi", dir: "BUY YES", price: 0.42 }, marketB: { title: "Will DOW hit 45K by July?", venue: "Polymarket", dir: "BUY NO", price: 0.38 }, currentRoi: 3.8, currentSpread: 4.0, costPerShare: 0.80, expiresIn: "9d", meetsCriteria: true, executed: true, estimatedProfit: 30.40 },
  { id: "pr-003", marketA: { title: "Fed cuts rates in Q3?", venue: "Kalshi", dir: "BUY YES", price: 0.65 }, marketB: { title: "Fed cuts rates in Q3?", venue: "Polymarket", dir: "BUY NO", price: 0.61 }, currentRoi: 3.5, currentSpread: 4.0, costPerShare: 1.26, expiresIn: "11d", meetsCriteria: true, executed: true, estimatedProfit: 44.10 },
  { id: "pr-004", marketA: { title: "S&P 500 > 5600 EOM?", venue: "Kalshi", dir: "BUY YES", price: 0.48 }, marketB: { title: "S&P 500 > 5600 EOM?", venue: "Polymarket", dir: "BUY NO", price: 0.44 }, currentRoi: 3.2, currentSpread: 4.0, costPerShare: 0.92, expiresIn: "16d", meetsCriteria: true, executed: false, estimatedProfit: 29.44 },
  { id: "pr-005", marketA: { title: "ETH ETF approved by Aug?", venue: "Kalshi", dir: "BUY YES", price: 0.55 }, marketB: { title: "ETH ETF approved by Aug?", venue: "Polymarket", dir: "BUY NO", price: 0.50 }, currentRoi: 5.0, currentSpread: 5.0, costPerShare: 1.05, expiresIn: "38d", meetsCriteria: true, executed: true, estimatedProfit: 52.50 },
  { id: "pr-006", marketA: { title: "Apple > $250 by Sept?", venue: "Kalshi", dir: "BUY YES", price: 0.37 }, marketB: { title: "Apple > $250 by Sept?", venue: "Polymarket", dir: "BUY NO", price: 0.33 }, currentRoi: 4.0, currentSpread: 4.0, costPerShare: 0.70, expiresIn: "83d", meetsCriteria: true, executed: false, estimatedProfit: 28.00 },
  { id: "pr-007", marketA: { title: "Lakers win NBA Finals 2026", venue: "Kalshi", dir: "BUY YES", price: 0.28 }, marketB: { title: "Lakers win NBA Finals 2026", venue: "Polymarket", dir: "BUY NO", price: 0.24 }, currentRoi: 4.5, currentSpread: 4.0, costPerShare: 0.52, expiresIn: "14d", meetsCriteria: true, executed: true, estimatedProfit: 23.40 },
  { id: "pr-008", marketA: { title: "Chiefs win Super Bowl LXI", venue: "Kalshi", dir: "BUY YES", price: 0.35 }, marketB: { title: "Chiefs win Super Bowl LXI", venue: "Polymarket", dir: "BUY NO", price: 0.31 }, currentRoi: 4.2, currentSpread: 4.0, costPerShare: 0.66, expiresIn: "221d", meetsCriteria: true, executed: false, estimatedProfit: 27.72 },
  { id: "pr-009", marketA: { title: "US GDP > 2% 2026?", venue: "Kalshi", dir: "BUY NO", price: 0.42 }, marketB: { title: "US GDP > 2% 2026?", venue: "Polymarket", dir: "BUY YES", price: 0.58 }, currentRoi: 2.8, currentSpread: 3.0, costPerShare: 1.00, expiresIn: "98d", meetsCriteria: false, executed: false, estimatedProfit: 28.00 },
  { id: "pr-010", marketA: { title: "Oil > $90 by Dec?", venue: "Kalshi", dir: "BUY YES", price: 0.35 }, marketB: { title: "Oil > $90 by Dec?", venue: "Polymarket", dir: "BUY NO", price: 0.31 }, currentRoi: 4.0, currentSpread: 4.0, costPerShare: 0.66, expiresIn: "160d", meetsCriteria: true, executed: true, estimatedProfit: 26.40 },
  { id: "pr-011", marketA: { title: "TSLA > $500 by Dec?", venue: "Kalshi", dir: "BUY NO", price: 0.55 }, marketB: { title: "TSLA > $500 by Dec?", venue: "Polymarket", dir: "BUY YES", price: 0.45 }, currentRoi: 3.7, currentSpread: 3.0, costPerShare: 1.00, expiresIn: "174d", meetsCriteria: true, executed: false, estimatedProfit: 37.00 },
  { id: "pr-012", marketA: { title: "Barcelona win La Liga 2026/27", venue: "Kalshi", dir: "BUY YES", price: 0.22 }, marketB: { title: "Barcelona win La Liga 2026/27", venue: "Polymarket", dir: "BUY NO", price: 0.19 }, currentRoi: 5.2, currentSpread: 3.0, costPerShare: 0.41, expiresIn: "300d", meetsCriteria: false, executed: false, estimatedProfit: 21.32 },
]

function filterPairs(type: string): MonitoredPair[] {
  switch (type) {
    case "Sports":
      return basePairs.filter(p => ["pr-007", "pr-008", "pr-012"].includes(p.id)).map(p => ({
        ...p, meetsCriteria: p.id !== "pr-012", executed: p.id === "pr-007",
      }))
    case "Commodities":
      return basePairs.filter(p => ["pr-010"].includes(p.id))
    case "Macro":
      return basePairs.filter(p => ["pr-003", "pr-009"].includes(p.id)).map(p => ({
        ...p, meetsCriteria: p.id === "pr-003", executed: p.id === "pr-003",
      }))
    case "Event":
      return basePairs.filter(p => ["pr-005", "pr-011"].includes(p.id)).map(p => ({
        ...p, meetsCriteria: true, executed: p.id === "pr-005",
      }))
    default:
      return basePairs.slice(0, 6).map(p => ({
        ...p, meetsCriteria: true, executed: Math.random() > 0.5,
      }))
  }
}

const d: StrategyConfig = { minRoi: 3.0, maxPositionsPerBundle: 1, maxPositionDollars: 1000, minSpread: 2.0, maxDailyExposure: 5000, cooldownMinutes: 30, autoExecute: true, notifyOnMatch: true, maxDaysToResolution: 90, enabled: true, paperMode: false }

export const strategyDetails: StrategyDetail[] = [
  { id: "strat-sprd-capture", name: "Spread Capture v2", type: "Spread", description: "Captures price discrepancies between Kalshi and Polymarket on identical event markets. Our flagship strategy.", status: "active", config: { ...d, minRoi: 2.5, maxDailyExposure: 8000 }, stats: { totalBundles: 64, completedBundles: 56, abortedBundles: 5, totalPnL: 18720, winRate: 87 }, monitoredPairs: filterPairs("Spread") },
  { id: "strat-event", name: "Event Driven", type: "Event", description: "Trades around specific binary events: earnings reports, FDA approvals, regulatory decisions. Higher ROI targets.", status: "active", config: { ...d, minRoi: 4.0, maxPositionDollars: 1500, maxDaysToResolution: 60 }, stats: { totalBundles: 38, completedBundles: 35, abortedBundles: 2, totalPnL: 18400, winRate: 92 }, monitoredPairs: filterPairs("Event") },
  { id: "strat-macro", name: "Macro Arbitrage", type: "Macro", description: "Broader economic indicator trading: GDP, inflation, Fed funds rate. Longer time horizons, lower volume.", status: "active", config: { ...d, minRoi: 2.0, maxDaysToResolution: 180, cooldownMinutes: 60 }, stats: { totalBundles: 18, completedBundles: 14, abortedBundles: 3, totalPnL: 6200, winRate: 78 }, monitoredPairs: filterPairs("Macro") },
  { id: "strat-commodity", name: "Commodity Arb", type: "Commodities", description: "Oil, gold, and commodity price prediction arbitrage. Medium confidence, wider spreads.", status: "active", config: { ...d, minRoi: 3.0, minSpread: 3.0, maxDailyExposure: 3000 }, stats: { totalBundles: 21, completedBundles: 13, abortedBundles: 6, totalPnL: 3240, winRate: 62 }, monitoredPairs: filterPairs("Commodities") },
  { id: "strat-sports", name: "Sports Arbitrage", type: "Sports", description: "Cross-venue sports betting arbitrage. NBA, NFL, Soccer. Fast resolution, high frequency.", status: "active", config: { ...d, minRoi: 3.5, maxDailyExposure: 6000, cooldownMinutes: 15, maxDaysToResolution: 14 }, stats: { totalBundles: 42, completedBundles: 38, abortedBundles: 3, totalPnL: 12500, winRate: 90 }, monitoredPairs: filterPairs("Sports") },
  { id: "strat-classic", name: "Classic Arbitrage", type: "Spread", description: "Original legacy strategy. Being phased out in favor of v2.", status: "paused", pausedReason: "Being deprecated — all pairs migrated to Spread Capture v2", config: { ...d, enabled: false, autoExecute: false }, stats: { totalBundles: 112, completedBundles: 89, abortedBundles: 18, totalPnL: 15200, winRate: 79 }, monitoredPairs: [] },
  { id: "strat-experimental", name: "ML Correlation Finder", type: "Event", description: "Experimental strategy using ML to find non-obvious correlated markets. Paper only.", status: "disabled", config: { ...d, autoExecute: false, paperMode: true, enabled: false, maxDailyExposure: 1000 }, stats: { totalBundles: 8, completedBundles: 5, abortedBundles: 3, totalPnL: 850, winRate: 62 }, monitoredPairs: filterPairs("Event").slice(0, 3) },
]
