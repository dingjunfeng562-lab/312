import { Button } from "./ui/button.tsx";
import { useConversationActions, useMessages } from "@/store/chat.ts";
import { MessageSquarePlus } from "lucide-react";

function ProjectLink() {
  const messages = useMessages();
  const { toggle } = useConversationActions();

  if (messages.length === 0) return null;

  return (
    <Button
      variant="outline"
      size="icon-md"
      className="rounded-full overflow-hidden"
      onClick={async () => await toggle(-1)}
    >
      <MessageSquarePlus className={`h-4 w-4`} />
    </Button>
  );
}

export default ProjectLink;
