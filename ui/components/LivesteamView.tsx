"use client";

import { users } from "@/fakedata/leftbar";
import { cn } from "@/utils/cn";
import { ClassValue } from "clsx";
import { ReactNode } from "react";
import stream_img from "@/public/images/stream_thumbnail_example.jpg";
import user_img from "@/public/images/user.jpg";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { LuMoreVertical } from "react-icons/lu";
import IconButton from "@/components/buttons/IconBtn";
import TagButton from "@/components/buttons/TagBtn";
import { Hover3DBox } from "@/components/Hover3DBox";
import { Stream } from "@/models/Stream";

const ContentView = ({
  title,
  channel,
  category,
  tags,
}: {
  title: string;
  channel: string;
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
      src={user_img}
      alt="mrbeast"
    />
      <div className="flex-1 flex-col space-y-1">
        <div className="w-full flex flex-row items-center justify-between font-semibold">
          <span className="text-sm hover:text-primary cursor-pointer">
            {title}
          </span>

          <IconButton icon={<LuMoreVertical className="w-4 h-4" />} />
        </div>
        <div className="text-sm text-secondaryWord cursor-pointer">
          {channel}
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

const LiveChannelView = ({
  className,
  viewers,
  title,
  category,
  tags,
  channel,
  onClick,
}: {
  className?: ClassValue;
  viewers: number;
  title: string;
  tags: string[];
  category?: string;
  channel: string;
  onClick?: () => void;
}) => {
  return (
    <div className={cn("flex flex-col gap-2", className)}>
      <Hover3DBox
        viewers={viewers}
        showViewer={true}
        showStream={true}
        imageSrc={stream_img}
        className="h-[170px]"
        onClick={onClick}
      />
      <ContentView
        channel={channel}
        title={title}
        category={category}
        tags={tags}
      />
    </div>
  );
};

const LiveChannelListView = ({
  limitView,
  streams,
}: {
  limitView: number;
  streams: Stream[];
}) => {
  const router = useRouter();
  streams = streams.slice(0, limitView);

  return (
    <div className="w-full grid xl:grid-cols-4 md:grid-cols-3 sm:grid-cols-2 max-sm:grid-cols-1 gap-4">
      {streams.map((stream, idx) => {
        const user = users.find((user) => user.id === stream.ownerId);
        return (
          <LiveChannelView
            key={idx}
            channel={user ? user.username : ""}
            title={stream.title}
            tags={stream.tags}
            viewers={120}
            category={stream.category}
            onClick={() => router.push(`/livestream`)}
          />
        );
      })}
    </div>
  );
};

const RecommendStreamView = ({
  title,
  streams,
  limitView = 4,
  separate,
}: {
  title: ReactNode;
  streams: Stream[];
  limitView?: number;
  separate: ReactNode;
}) => {
  return (
    <div className="flex flex-col gap-2 mt-8 pr-2">
      <div className="font-semibold text-lg">{title}</div>
      <LiveChannelListView limitView={limitView} streams={streams} />
      {separate}
    </div>
  );
};

export {
  ContentView,
  LiveChannelListView,
  LiveChannelView,
  RecommendStreamView,
};