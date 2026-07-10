import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  deleteInvitation,
  disableInvitation,
  enableInvitation,
  generateInvitation,
  generateInvitationAdvanced,
  getInvitationList,
  getInvitationUsage,
  InvitationCode,
  InvitationUsageDetail,
} from "@/admin/api/invitation";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Ban, CheckCircle, Copy, Eye, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";

type InvitationManagementProps = {
  embedded?: boolean;
};

function InvitationManagement({ embedded = false }: InvitationManagementProps) {
  const { t } = useTranslation();
  const [invitations, setInvitations] = useState<InvitationCode[]>([]);
  const [page, setPage] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [generateOpen, setGenerateOpen] = useState(false);
  const [advancedMode, setAdvancedMode] = useState(false);
  const [usageOpen, setUsageOpen] = useState(false);
  const [usageLoading, setUsageLoading] = useState(false);
  const [usageDetail, setUsageDetail] =
    useState<InvitationUsageDetail | null>(null);
  const [formData, setFormData] = useState({
    type: "AI",
    quota: 10,
    number: 1,
    expires_days: 0,
    notes: "",
  });

  const loadInvitations = async () => {
    setLoading(true);
    const resp = await getInvitationList(page);
    if (resp.status) {
      setInvitations(resp.data);
      setTotal(resp.total);
    } else {
      toast.error(t("error"), {
        description: resp.message || "Failed to load invitations",
      });
    }
    setLoading(false);
  };

  useEffect(() => {
    loadInvitations();
  }, [page]);

  const handleGenerate = async () => {
    if (formData.number < 1 || formData.number > 100) {
      toast.error(t("error"), {
        description: "Number must be between 1 and 100",
      });
      return;
    }

    const data = advancedMode
      ? await generateInvitationAdvanced(formData)
      : await generateInvitation({
          type: formData.type,
          quota: formData.quota,
          number: formData.number,
        });

    if (data.status) {
      toast.success(t("success"), {
        description: `Generated ${data.data?.length} invitation codes`,
      });
      setGenerateOpen(false);
      loadInvitations();

      if (data.data && data.data.length > 0) {
        const codes = data.data.join("\n");
        navigator.clipboard.writeText(codes);
        toast.success("Copied to clipboard", {
          description: `${data.data.length} codes copied`,
        });
      }
    } else {
      toast.error(t("error"), {
        description: data.message || "Failed to generate",
      });
    }
  };

  const handleViewUsage = async (code: string) => {
    setUsageOpen(true);
    setUsageLoading(true);
    setUsageDetail(null);

    const resp = await getInvitationUsage(code);
    if (resp.status && resp.data) {
      setUsageDetail(resp.data);
    } else {
      toast.error(t("error"), {
        description: resp.message || "Failed to load invitation detail",
      });
    }

    setUsageLoading(false);
  };

  const handleCopy = (code: string) => {
    navigator.clipboard.writeText(code);
    toast.success("Copied", {
      description: `Code ${code} copied to clipboard`,
    });
  };

  const handleDelete = async (code: string) => {
    if (!confirm(`Are you sure to delete invitation code: ${code}?`)) {
      return;
    }

    const resp = await deleteInvitation(code);
    if (resp.status) {
      toast.success(t("success"), {
        description: "Invitation code deleted",
      });
      loadInvitations();
    } else {
      toast.error(t("error"), {
        description: resp.error || "Failed to delete",
      });
    }
  };

  const handleDisable = async (code: string) => {
    const resp = await disableInvitation(code);
    if (resp.status) {
      toast.success(t("success"), {
        description: "Invitation code disabled",
      });
      loadInvitations();
    } else {
      toast.error(t("error"), {
        description: resp.error || "Failed to disable",
      });
    }
  };

  const handleEnable = async (code: string) => {
    const resp = await enableInvitation(code);
    if (resp.status) {
      toast.success(t("success"), {
        description: "Invitation code enabled",
      });
      loadInvitations();
    } else {
      toast.error(t("error"), {
        description: resp.error || "Failed to enable",
      });
    }
  };

  const getStatusBadge = (invitation: InvitationCode) => {
    if (invitation.used) {
      return <Badge variant="secondary">Used</Badge>;
    }
    if (invitation.is_expired) {
      return <Badge variant="destructive">Expired</Badge>;
    }
    if (invitation.status === "disabled") {
      return <Badge variant="outline">Disabled</Badge>;
    }
    return <Badge variant="default">Unused</Badge>;
  };

  const formatDate = (value?: string | null) => {
    if (!value) {
      return "-";
    }

    const date = new Date(value.replace(" ", "T"));
    return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
  };

  const detailRows: [string, string | number][] = usageDetail
    ? [
        ["Code", usageDetail.code],
        ["Quota", usageDetail.quota],
        ["Type", usageDetail.type],
        [
          "Status",
          usageDetail.used
            ? "Used"
            : usageDetail.is_expired
              ? "Expired"
              : "Unused",
        ],
        ["Created By", usageDetail.creator_name || "system"],
        ["Created At", formatDate(usageDetail.created_at)],
        ["Expires At", formatDate(usageDetail.expires_at)],
        ["Used By", usageDetail.used_by_user || "-"],
        ["Used At", formatDate(usageDetail.used_at)],
        ["Used IP", usageDetail.used_ip || "-"],
        ["Notes", usageDetail.notes || "-"],
      ]
    : [];

  return (
    <div
      className={
        embedded ? "invitation-management" : "invitation-management p-6"
      }
    >
      <div className="mb-6 flex items-center justify-between gap-4">
        {!embedded && (
          <div>
            <h2 className="text-2xl font-bold">
              Invitation Code Management
            </h2>
            <p className="mt-1 text-muted-foreground">
              Manage and generate invitation codes for user registration
            </p>
          </div>
        )}
        <Dialog open={generateOpen} onOpenChange={setGenerateOpen}>
          <DialogTrigger asChild>
            <Button className={embedded ? "ml-auto" : undefined}>
              <Plus className="mr-2 h-4 w-4" />
              Generate Codes
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle>Generate Invitation Codes</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  checked={advancedMode}
                  onChange={(e) => setAdvancedMode(e.target.checked)}
                  id="advanced"
                />
                <Label htmlFor="advanced">Advanced Mode (expiry & notes)</Label>
              </div>

              <div className="space-y-2">
                <Label>Type (Prefix)</Label>
                <Input
                  value={formData.type}
                  onChange={(e) =>
                    setFormData({ ...formData, type: e.target.value })
                  }
                  placeholder="AI, VIP, etc."
                />
              </div>

              <div className="space-y-2">
                <Label>Quota (per code)</Label>
                <Input
                  type="number"
                  value={formData.quota}
                  onChange={(e) =>
                    setFormData({ ...formData, quota: Number(e.target.value) })
                  }
                  min="0"
                />
              </div>

              <div className="space-y-2">
                <Label>Number of Codes</Label>
                <Input
                  type="number"
                  value={formData.number}
                  onChange={(e) =>
                    setFormData({ ...formData, number: Number(e.target.value) })
                  }
                  min="1"
                  max="100"
                />
              </div>

              {advancedMode && (
                <>
                  <div className="space-y-2">
                    <Label>Expires in (days, 0 = never)</Label>
                    <Input
                      type="number"
                      value={formData.expires_days}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          expires_days: Number(e.target.value),
                        })
                      }
                      min="0"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label>Notes</Label>
                    <Input
                      value={formData.notes}
                      onChange={(e) =>
                        setFormData({ ...formData, notes: e.target.value })
                      }
                      placeholder="Optional description"
                    />
                  </div>
                </>
              )}

              <Button className="w-full" onClick={handleGenerate}>
                Generate
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Code</TableHead>
              <TableHead>Quota</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Used By</TableHead>
              <TableHead>Creator</TableHead>
              <TableHead>Created At</TableHead>
              <TableHead>Expires At</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && (
              <TableRow>
                <TableCell
                  colSpan={8}
                  className="text-center text-muted-foreground"
                >
                  Loading...
                </TableCell>
              </TableRow>
            )}
            {!loading && invitations.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={8}
                  className="text-center text-muted-foreground"
                >
                  No invitation codes
                </TableCell>
              </TableRow>
            )}
            {!loading &&
              invitations.map((inv) => (
                <TableRow key={inv.code}>
                  <TableCell className="font-mono text-sm">
                    {inv.code}
                  </TableCell>
                  <TableCell>{inv.quota}</TableCell>
                  <TableCell>{getStatusBadge(inv)}</TableCell>
                  <TableCell>{inv.username || "-"}</TableCell>
                  <TableCell>{inv.creator_name || "system"}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(inv.created_at)}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(inv.expires_at)}
                  </TableCell>
                  <TableCell>
                    <div className="flex space-x-2">
                      <Button
                        size="sm"
                        variant="ghost"
                        title="View usage detail"
                        onClick={() => handleViewUsage(inv.code)}
                      >
                        <Eye className="h-4 w-4" />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        title="Copy"
                        onClick={() => handleCopy(inv.code)}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                      {!inv.used && inv.status !== "disabled" && (
                        <Button
                          size="sm"
                          variant="ghost"
                          title="Disable"
                          onClick={() => handleDisable(inv.code)}
                        >
                          <Ban className="h-4 w-4" />
                        </Button>
                      )}
                      {!inv.used && inv.status === "disabled" && (
                        <Button
                          size="sm"
                          variant="ghost"
                          title="Enable"
                          onClick={() => handleEnable(inv.code)}
                        >
                          <CheckCircle className="h-4 w-4" />
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        title="Delete"
                        onClick={() => handleDelete(inv.code)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
          </TableBody>
        </Table>
      </div>

      <div className="mt-4 flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          Page {page + 1} of {total}
        </div>
        <div className="flex space-x-2">
          <Button
            variant="outline"
            onClick={() => setPage(Math.max(0, page - 1))}
            disabled={page === 0}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            onClick={() => setPage(page + 1)}
            disabled={page >= total - 1}
          >
            Next
          </Button>
        </div>
      </div>

      <Dialog open={usageOpen} onOpenChange={setUsageOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Invitation Usage Detail</DialogTitle>
          </DialogHeader>
          {usageLoading ? (
            <div className="py-6 text-center text-muted-foreground">
              Loading...
            </div>
          ) : usageDetail ? (
            <div className="grid gap-3 py-2">
              {detailRows.map(([label, value]) => (
                <div
                  key={label}
                  className="grid grid-cols-[120px_1fr] items-start gap-3 text-sm"
                >
                  <span className="text-muted-foreground">{label}</span>
                  <span
                    className={
                      label === "Code" ? "break-all font-mono" : "break-words"
                    }
                  >
                    {value}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-6 text-center text-muted-foreground">
              No detail available
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default InvitationManagement;
