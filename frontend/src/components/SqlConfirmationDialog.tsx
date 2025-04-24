// src/components/SqlConfirmationDialog.tsx
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { ConfirmationDialog } from "../types/query";

interface SqlConfirmationDialogProps {
  confirmationDialog: ConfirmationDialog;
  onConfirm: () => void;
  onCancel: () => void;
}

export default function SqlConfirmationDialog({
  confirmationDialog,
  onConfirm,
  onCancel,
}: SqlConfirmationDialogProps) {
  return (
    <Dialog
      open={confirmationDialog.open}
      onOpenChange={(open) => {
        if (!open) onCancel();
      }}
    >
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle className="flex items-center text-amber-600">
            <AlertTriangle className="h-5 w-5 mr-2" /> Confirmation Required
          </DialogTitle>
          <DialogDescription>
            You are about to execute a query that will modify your database.
            Please review the SQL before proceeding.
          </DialogDescription>
        </DialogHeader>

        <div className="bg-amber-50 p-4 rounded border border-amber-200 my-4 overflow-auto">
          <h4 className="font-semibold text-amber-800 mb-2">SQL Query:</h4>
          <pre className="bg-white p-3 rounded font-mono text-sm overflow-x-auto border border-amber-100">
            {confirmationDialog.sql}
          </pre>
        </div>

        <DialogFooter className="flex justify-between sm:justify-between">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            className="mt-2 sm:mt-0"
          >
            Cancel
          </Button>
          <Button
            type="button"
            className="bg-amber-600 hover:bg-amber-700 mt-2 sm:mt-0"
            onClick={onConfirm}
          >
            Confirm and Execute
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
