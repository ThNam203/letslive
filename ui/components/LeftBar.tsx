import { LuArrowUpDown, LuHeart, LuVideo } from "react-icons/lu";
import { ClassValue } from "clsx";
import { cn } from "@/utils/cn";
import { ReactNode } from "react";
import user_avatar from "@/public/images/user_avatar.jpeg";
import { Channel } from "@/types/Channel";
import { channels } from "@/fakedata/leftbar";
import IconButton from "@/components/buttons/IconBtn";
import RoundedImage from "@/components/images/RoundedImage";

const ChannelViewItem = ({
  className,
  name,
  category,
}: {
  className?: ClassValue;
  name: string;
  category: string;
}) => {
  return (
    <div
      className={cn(
        "w-full flex flex-row gap-2 items-center justify-between xl:hover:bg-hoverColor px-2",
        className
      )}
    >
      <div className="flex flex-row items-center justify-start">
        <RoundedImage src={user_avatar} width={40} height={40} alt="channel owner image"/>
        <div className="flex flex-col gap-1 ml-2 max-xl:hidden">
          <span className="font-semibold text-sm">{name}</span>
          <span className="text-secondaryWord text-sm">{category}</span>
        </div>
      </div>
      <View viewers={1200} className="max-xl:hidden" />
    </div>
  );
};

const View = ({
  className,
  viewers,
}: {
  className?: ClassValue;
  viewers: number;
}) => {
  return (
    <div
      className={cn(
        "flex flex-row items-center justify-center gap-2",
        className
      )}
    >
      <span className="w-2 h-2 rounded-full bg-red-600"></span>
      <span className="text-xs">{viewers.toString()}</span>
    </div>
  );
};

const recommendChannels: Channel[] = [channels[1], channels[2], channels[3]];

export const LeftBar = () => {
  return (
    <div className="fixed h-full xl:w-64 max-xl:w-fit bg-leftBarColor flex flex-col max-xl:items-center justify-start gap-2 py-2 text-primaryWord">
      <div className="flex flex-row justify-between items-center px-2">
        <span className="font-semibold text-lg max-xl:hidden">For you</span>
      </div>
      <div className="flex flex-row justify-between items-center px-2 mt-2">
        <Title icon={<LuHeart size={20} />}>FOLLOWED CHANNELS</Title>

        <IconButton
          className="self-end max-xl:hidden"
          icon={<LuArrowUpDown size={18} />}
          disabled={true}
        />
      </div>
      {/* {followingChannels.map((channel, idx) => {
        // TODO: WRONG LOGIC
        const user = userData.find((user) => user.id === channel.id);
        return (
          <ChannelViewItem
            key={idx}
            name={user ? user.username : ""}
            category={"what is this"}
          />
        );
      })} */}
      <div className="flex flex-row justify-between items-center px-2 mt-2">
        <Title icon={<LuVideo size={20} />}>RECOMMEND CHANNELS</Title>
      </div>
      {recommendChannels.map((channel, idx) => {
        return (
          <ChannelViewItem
            key={idx}
            name={"CHANNEL OWNER"}
            category={"CATEGORY"}
          />
        );
      })}
    </div>
  );
};

const Title = ({
  children,
  icon,
}: {
  children: ReactNode;
  icon: ReactNode;
}) => {
  return (
    <>
      <span className="font-semibold text-sm text-secondaryWord max-xl:hidden">
        {children}
      </span>
      <span className="font-semibold text-sm text-secondaryWord xl:hidden">
        {icon}
      </span>
    </>
  );
};