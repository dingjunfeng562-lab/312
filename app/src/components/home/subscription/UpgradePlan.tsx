import { Button } from "@/components/ui/button.tsx";
import { useTranslation } from "react-i18next";

// 订阅功能已移除，保留此组件以保持兼容性

type UpgradeProps = {
  level: number;
  current: number;
  isYearly?: boolean;
};

export function Upgrade(_props: UpgradeProps) {
  const { t } = useTranslation();

  return (
    <Button disabled className="action w-full" variant="outline">
      {t("Subscription disabled")}
    </Button>
  );
}
