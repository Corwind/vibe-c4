interface BreadcrumbEntry {
  label: string;
}

interface BreadcrumbsProps {
  items: BreadcrumbEntry[];
  onNavigate: (index: number) => void;
}

export function Breadcrumbs({ items, onNavigate }: BreadcrumbsProps) {
  return (
    <nav aria-label="Diagram breadcrumb" className="flex items-center gap-1 text-sm">
      {items.map((item, index) => {
        const isLast = index === items.length - 1;
        return (
          <span key={index} className="flex items-center gap-1">
            {index > 0 && (
              <span className="text-gray-400" aria-hidden="true">
                /
              </span>
            )}
            {isLast ? (
              <span className="font-medium text-text" aria-current="page">
                {item.label}
              </span>
            ) : (
              <button
                onClick={() => onNavigate(index)}
                className="text-primary hover:underline"
              >
                {item.label}
              </button>
            )}
          </span>
        );
      })}
    </nav>
  );
}
