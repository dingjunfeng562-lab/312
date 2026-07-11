import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card.tsx";
import { Button } from "@/components/ui/button.tsx";
import InvitationTable from "@/components/admin/InvitationTable.tsx";
import RedeemTable from "@/components/admin/RedeemTable.tsx";
import router from "@/router.tsx";
import { Gift, TicketCheck } from "lucide-react";
import { useTranslation } from "react-i18next";

type CodeManagementProps = {
  type: "invitation" | "gift";
};

function CodeManagement({ type }: CodeManagementProps) {
  const { t } = useTranslation();
  const invitation = type === "invitation";
  const title = invitation
    ? t("code-management.invitation-title")
    : t("code-management.gift-title");
  const description = invitation
    ? t("code-management.invitation-description")
    : t("code-management.gift-description");

  return (
    <div className="user-interface">
      <Card className="admin-card">
        <CardHeader className="select-none gap-3 md:flex-row md:items-start md:justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              {invitation ? (
                <TicketCheck className="h-5 w-5" />
              ) : (
                <Gift className="h-5 w-5" />
              )}
              {title}
            </CardTitle>
            <CardDescription className="mt-2">{description}</CardDescription>
          </div>
          <Button
            variant="outline"
            onClick={() =>
              router.navigate(invitation ? "/admin/gift-code" : "/admin/invitation")
            }
          >
            {invitation ? (
              <Gift className="mr-2 h-4 w-4" />
            ) : (
              <TicketCheck className="mr-2 h-4 w-4" />
            )}
            {invitation
              ? t("code-management.go-to-gift")
              : t("code-management.go-to-invitation")}
          </Button>
        </CardHeader>
        <CardContent>
          {invitation ? <InvitationTable /> : <RedeemTable />}
        </CardContent>
      </Card>
    </div>
  );
}

export function InvitationCodeManagement() {
  return <CodeManagement type="invitation" />;
}

export function GiftCodeManagement() {
  return <CodeManagement type="gift" />;
}