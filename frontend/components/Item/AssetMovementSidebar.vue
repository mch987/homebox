<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Textarea } from "@/components/ui/textarea";
  import type { AssetMovementHistory } from "~~/lib/api/classes/assets";
  import type { EntityOut, EntitySummary, UserSummary } from "~~/lib/api/types/data-contracts";

  const props = defineProps<{ item: EntityOut }>();

  const api = useUserApi();

  const assetId = ref(props.item.assetId || "");
  const loadingAsset = ref(false);
  const assetDetails = ref<EntitySummary | null>(null);
  const history = ref<AssetMovementHistory[]>([]);

  const locations = ref<EntitySummary[]>([]);
  const users = ref<UserSummary[]>([]);

  const movementType = ref<"Transfer" | "Check Out" | "Check In" | "Return" | "Maintenance">("Transfer");
  const destinationLocationId = ref("");
  const movedByUserId = ref("");
  const notes = ref("");
  const otherPersonName = ref("");
  const otherPersonEmail = ref("");
  const otherPersonDepartment = ref("");
  const errors = ref<string[]>([]);
  const step = ref<"edit" | "review">("edit");

  const isOtherUser = computed(() => movedByUserId.value === "other");

  async function loadDropdowns() {
    const [locRes, memberRes] = await Promise.all([api.items.getLocations(), api.group.getMembers()]);
    if (!locRes.error) {
      locations.value = locRes.data;
    }
    if (!memberRes.error) {
      users.value = memberRes.data;
    }
  }

  async function lookupAsset() {
    errors.value = [];
    if (!assetId.value.trim()) {
      errors.value.push("Asset ID is required.");
      return;
    }

    loadingAsset.value = true;
    try {
      const res = await api.assets.get(assetId.value.trim());
      if (res.error || !res.data || res.data.total < 1) {
        assetDetails.value = null;
        history.value = [];
        errors.value.push("Invalid Asset ID.");
        return;
      }
      assetDetails.value = res.data.items[0] || null;
      const histRes = await api.assets.getMovementHistory(assetId.value.trim(), 8);
      history.value = histRes.error ? [] : histRes.data || [];
    } finally {
      loadingAsset.value = false;
    }
  }

  function validate() {
    const list: string[] = [];
    if (!assetId.value) list.push("Asset ID is required.");
    if (!assetDetails.value) list.push("Lookup a valid Asset ID before submitting.");
    if (!movementType.value) list.push("Movement type is required.");
    if (!destinationLocationId.value && movementType.value !== "Maintenance") {
      list.push("Destination location is required.");
    }
    if (!movedByUserId.value) list.push("Moved by is required.");
    if (isOtherUser.value && !otherPersonName.value.trim()) {
      list.push("Other person full name is required.");
    }
    errors.value = list;
    return list.length === 0;
  }

  function review() {
    if (!validate()) return;
    step.value = "review";
  }

  async function submit(overrideCheckedOutLock = false) {
    if (!validate()) return;

    const payload = {
      movementType: movementType.value,
      toLocationId: destinationLocationId.value || undefined,
      movedByUserId: isOtherUser.value ? undefined : movedByUserId.value,
      otherPersonName: isOtherUser.value ? otherPersonName.value : undefined,
      otherPersonEmail: isOtherUser.value ? otherPersonEmail.value : undefined,
      otherPersonDepartment: isOtherUser.value ? otherPersonDepartment.value : undefined,
      notes: notes.value || undefined,
      overrideCheckedOutLock,
    };

    const res = await api.assets.createMovement(assetId.value.trim(), payload);
    if (res.error) {
      const err = res.error as { data?: string; message?: string } | null;
      const errMsg = String(err?.data || err?.message || "Failed to move asset");
      if (errMsg.includes("override required")) {
        if (window.confirm("This asset appears to be checked out already. Override and continue?")) {
          await submit(true);
        }
        return;
      }
      errors.value = [errMsg];
      return;
    }

    toast.success("Asset movement saved.");
    step.value = "edit";
    notes.value = "";
    const histRes = await api.assets.getMovementHistory(assetId.value.trim(), 8);
    history.value = histRes.error ? history.value : histRes.data || [];
  }

  onMounted(async () => {
    await loadDropdowns();
    await lookupAsset();
  });
</script>

<template>
  <div class="space-y-4">
    <div class="grid gap-2">
      <Label>Asset ID</Label>
      <div class="flex gap-2">
        <Input v-model="assetId" placeholder="Enter or scan Asset ID" @keyup.enter="lookupAsset" />
        <Button type="button" variant="outline" :disabled="loadingAsset" @click="lookupAsset">Lookup</Button>
      </div>
    </div>

    <div v-if="assetDetails" class="rounded-md border p-3 text-sm">
      <p><strong>Name:</strong> {{ assetDetails.name }}</p>
      <p><strong>Description:</strong> {{ assetDetails.description || "-" }}</p>
      <p><strong>Current location:</strong> {{ assetDetails.parent?.name || "-" }}</p>
      <p><strong>Current status:</strong> {{ history[0]?.newStatus || "Available" }}</p>
      <p><strong>Assigned user:</strong> {{ history[0]?.otherPersonName || "-" }}</p>
    </div>

    <div class="grid gap-2">
      <Label>Movement type</Label>
      <Select v-model="movementType">
        <SelectTrigger><SelectValue /></SelectTrigger>
        <SelectContent>
          <SelectItem value="Transfer">Transfer</SelectItem>
          <SelectItem value="Check Out">Check Out</SelectItem>
          <SelectItem value="Check In">Check In</SelectItem>
          <SelectItem value="Return">Return</SelectItem>
          <SelectItem value="Maintenance">Maintenance</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="grid gap-2">
      <Label>Destination location</Label>
      <Select v-model="destinationLocationId">
        <SelectTrigger><SelectValue placeholder="Select destination" /></SelectTrigger>
        <SelectContent>
          <SelectItem v-for="loc in locations" :key="loc.id" :value="loc.id">{{ loc.name }}</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="grid gap-2">
      <Label>Moved by</Label>
      <Select v-model="movedByUserId">
        <SelectTrigger><SelectValue placeholder="Select user" /></SelectTrigger>
        <SelectContent>
          <SelectItem v-for="u in users" :key="u.id" :value="u.id">{{ u.name }}</SelectItem>
          <SelectItem value="other">Other</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <template v-if="isOtherUser">
      <div class="grid gap-2">
        <Label>Person full name</Label>
        <Input v-model="otherPersonName" />
      </div>
      <div class="grid gap-2">
        <Label>Email or contact</Label>
        <Input v-model="otherPersonEmail" />
      </div>
      <div class="grid gap-2">
        <Label>Department or notes</Label>
        <Input v-model="otherPersonDepartment" />
      </div>
    </template>

    <div class="grid gap-2">
      <Label>Notes/comments (optional)</Label>
      <Textarea v-model="notes" />
    </div>

    <div
      v-if="errors.length"
      class="rounded-md border border-destructive/60 bg-destructive/10 p-3 text-sm text-destructive"
    >
      <ul class="list-disc pl-5">
        <li v-for="e in errors" :key="e">{{ e }}</li>
      </ul>
    </div>

    <div class="flex gap-2">
      <Button v-if="step === 'edit'" @click="review">Review</Button>
      <template v-else>
        <Button @click="submit">Confirm Submit</Button>
        <Button variant="outline" @click="step = 'edit'">Back</Button>
      </template>
    </div>

    <div v-if="history.length" class="pt-4">
      <h3 class="mb-2 text-sm font-semibold">Recent movement history</h3>
      <table class="w-full text-left text-xs">
        <thead>
          <tr class="border-b">
            <th class="py-1">Date</th>
            <th class="py-1">Type</th>
            <th class="py-1">From</th>
            <th class="py-1">To</th>
            <th class="py-1">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="h in history" :key="h.movementHistoryId" class="border-b">
            <td class="py-1">{{ new Date(h.movementDateTime).toLocaleString() }}</td>
            <td class="py-1">{{ h.movementType }}</td>
            <td class="py-1">{{ h.fromLocationId || "-" }}</td>
            <td class="py-1">{{ h.toLocationId || "-" }}</td>
            <td class="py-1">{{ h.newStatus || "-" }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
