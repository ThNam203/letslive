import Image from "next/image";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/utils/cn";

const SIZES = {
    lg: {
        card: "gap-2 rounded-xl p-4",
        image: "h-24 w-24",
        name: "text-sm font-semibold",
    },
    md: {
        card: "gap-2 rounded-xl p-4",
        image: "h-16 w-16",
        name: "text-sm font-medium",
    },
    sm: {
        card: "gap-1 rounded-lg p-3",
        image: "h-16 w-16",
        name: "text-xs font-medium",
    },
} as const;

/**
 * The image + name + badge card shared by the shop, inventory, received gifts
 * and the gift picker. With `onClick` the whole card is a button; otherwise
 * `children` can hold an action or extra text below the badge.
 */
export default function ShopItemCard({
    name,
    imageUrl,
    badge,
    description,
    size = "lg",
    onClick,
    disabled,
    children,
}: {
    name: string;
    /** omitted when the item is no longer in the catalogue */
    imageUrl?: string;
    badge: string;
    description?: string | null;
    size?: keyof typeof SIZES;
    onClick?: () => void;
    disabled?: boolean;
    children?: React.ReactNode;
}) {
    const styles = SIZES[size];

    const content = (
        <>
            {imageUrl && (
                <div className={cn("relative", styles.image)}>
                    <Image
                        src={imageUrl}
                        alt={name}
                        fill
                        className="object-contain"
                        unoptimized
                    />
                </div>
            )}
            <p className={cn("text-foreground text-center", styles.name)}>
                {name}
            </p>
            {description && (
                <p className="text-muted-foreground line-clamp-2 text-center text-xs">
                    {description}
                </p>
            )}
            <Badge variant="secondary">{badge}</Badge>
            {children}
        </>
    );

    const base = cn(
        "border-border bg-card flex flex-col items-center border",
        styles.card,
    );

    return onClick ? (
        <button
            type="button"
            onClick={onClick}
            disabled={disabled}
            className={cn(
                base,
                "hover:border-primary transition-colors disabled:opacity-50",
            )}
        >
            {content}
        </button>
    ) : (
        <div className={base}>{content}</div>
    );
}
