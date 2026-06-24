export function StatCard({ label, value, sub, positive, negative, accent }: {
  label: string; value: string; sub?: string; positive?: boolean; negative?: boolean; accent?: boolean
}) {
  return (
    <div className="bg-surface-alt rounded-xl border border-border p-4">
      <p className="text-xs text-muted mb-0.5">{label}</p>
      <p className="text-2xl font-bold text-gray-100">{value}</p>
      {sub && (
        <p className={`text-sm mt-0.5 ${accent ? "text-accent" : positive ? "text-green" : negative ? "text-red" : "text-muted"}`}>
          {sub}
        </p>
      )}
    </div>
  )
}
