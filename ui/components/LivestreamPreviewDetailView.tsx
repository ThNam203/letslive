import { cn } from "@/utils/cn";
import user_avatar from "@/public/images/user_avatar.jpeg";
import Image from "next/image";
import { LuMoreVertical } from "react-icons/lu";
import IconButton from "@/components/buttons/IconBtn";
import TagButton from "@/components/buttons/TagBtn";

const LivestreamPreviewDetailView = ({
    title,
    username,
    category,
    tags,
}: {
    title: string;
    username: string;
    category: string | undefined;
    tags: string[];
}) => {
    return (
        <div className="flex flex-row gap-2">
            <Image
                width={500}
                height={500}
                className={cn(
                    "h-8 w-8 rounded-full overflow-hidden cursor-pointer"
                )}
                src={user_avatar}
                alt="user avatar"
            />
            <div className="flex-1 flex-col space-y-1">
                <div className="w-full flex flex-row items-center justify-between font-semibold">
                    <span className="text-sm hover:text-primary cursor-pointer">
                        {title}
                    </span>

                    <IconButton icon={<LuMoreVertical className="w-4 h-4" />} />
                </div>
                <div className="text-sm text-secondaryWord cursor-pointer">
                    {username}
                </div>
                <div className="text-sm text-secondaryWord hover:text-primary cursor-pointer">
                    {category ? category : null}
                </div>
                <div className="flex flex-row gap-2 justify-self-end">
                    {tags.map((tag, idx) => {
                        return <TagButton key={idx} content={tag} />;
                    })}
                </div>
            </div>
        </div>
    );
};

export default LivestreamPreviewDetailView;