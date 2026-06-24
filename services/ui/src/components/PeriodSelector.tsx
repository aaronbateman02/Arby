interface Props {
  periods: string[]
  selected: string
  onSelect: (p: string) => void
}

export function PeriodSelector({ periods, selected, onSelect }: Props) {
  return (
    <div className="flex gap-1 bg-surface-alt rounded-lg p-1 border border-border w-fit">
      {periods.map((p) => (
        <button
          key={p}
          onClick={() => onSelect(p)}
          className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
            p === selected ? "bg-surface text-gray-100 font-medium shadow-sm border border-border" : "text-muted hover:text-gray-200"
          }`}
        >
          {p}
        </button>
      ))}
    </div>
  )
}
