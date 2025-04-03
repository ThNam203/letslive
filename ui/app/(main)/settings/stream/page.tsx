"use client";

import { Loader } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { toast } from "react-toastify";
import { Button } from "../../../../components/ui/button";
import useUser from "../../../../hooks/user";
import { UpdateLivestreamInformation } from "../../../../lib/api/user";
import ImageField from "../_components/image-field";
import Section from "../_components/section";
import TextField from "../_components/text-field";
import TextAreaField from "../_components/textarea-field";

export default function StreamEdit() {
  const user = useUser((state) => state.user);
  const updateUser = useUser((state) => state.updateUser);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [image, setImage] = useState<File | null>(null);
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleImageChange = (file: File | null) => {
    if (file) {
      const imageUrl = URL.createObjectURL(file);
      setImage(file);
      setImageUrl(imageUrl);
    }
  };

  const handleResetImage = () => {
    setImage(null);
    setImageUrl(null);
  };

  useEffect(() => {
    if (user) {
      setTitle(user.livestreamInformation.title || "");
      setDescription(user.livestreamInformation.description || "");
      setImageUrl(user.livestreamInformation.thumbnailUrl || null);
      setImage(null);
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;

    setIsSubmitting(true);
    const { updatedInfo, fetchError } = await UpdateLivestreamInformation(
      image,
      user!.livestreamInformation.thumbnailUrl,
      title,
      description
    );
    setIsSubmitting(false);
    if (fetchError) {
      toast(fetchError.message, { type: "error" });
      return;
    }

    if (updatedInfo) {
      toast.success("Livestream information updated successfully");
      updateUser({
        ...user,
        livestreamInformation: {
          userId: updatedInfo.userId,
          title: updatedInfo.title,
          description: updatedInfo.description,
          thumbnailUrl: updatedInfo.thumbnailUrl,
        },
      });
    }
  };

  const isFormChange = useMemo(() => {
    return (
      title !== user?.livestreamInformation.title ||
      description !== user?.livestreamInformation.description ||
      imageUrl !== user?.livestreamInformation.thumbnailUrl
    );
  }, [title, description, imageUrl, user]);

  return (
    <div className="min-h-screen max-w-4xl text-gray-900 p-6">
      <Section
        title="Livestream"
        description={`Your next livestream information will be based on the information.\nIt won't change even after livestream ends.`}
        hasBorder
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <TextField
            label="Title"
            description="If empty, the title will be generated automatically."
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
          />
          <TextAreaField
            label="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="resize-none"
            required
          />
          <ImageField
            label="Thumbnail"
            description="If empty, the thumbnail will be generated automatically."
            imageUrl={imageUrl}
            hoverText="Change thumbnail"
            onImageChange={handleImageChange}
            onResetImage={handleResetImage}
            showCloseIcon={imageUrl !== null}
          />

          <div className="flex justify-end items-center">
            <Button
              className="disabled:bg-gray-200 disabled:hover:cursor-not-allowed"
              disabled={isSubmitting || !isFormChange}
              type="submit"
            >
              {isSubmitting && <Loader className="animate-spin" />}
              Confirm edit
            </Button>
          </div>
        </form>
      </Section>
    </div>
  );
}
