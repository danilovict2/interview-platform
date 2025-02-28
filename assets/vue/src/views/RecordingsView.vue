<template>
    <AppLoading v-if="isLoading" />
    <div class="container max-w-7xl mx-auto p-6" v-else>
        <h1 class="text-3xl font-bold">Recordings</h1>
        <p class="text-muted-foreground my-1">
            {{ recordings.length }}
            {{ recordings.length === 1 ? 'recording' : 'recordings' }} available
        </p>

        <ScrollArea class="h-[calc(100vh-12rem)] mt-3">
            <div
                class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6 pb-6"
                v-if="recordings.length"
            >
                <RecordingCard
                    v-for="(recording, index) in recordings"
                    :key="index"
                    :recording="recording"
                />
            </div>
            <div class="flex flex-col items-center justify-center h-[400px] gap-4" v-else>
                <p class="text-xl font-medium text-muted-foreground">No recordings available</p>
            </div>
        </ScrollArea>
    </div>
</template>

<script setup>
import AppLoading from '@/components/AppLoading.vue';
import RecordingCard from '@/components/RecordingCard.vue';
import ScrollArea from '@/components/ui/scroll-area/ScrollArea.vue';
import router from '@/router';
import { useAuthStore } from '@/stores/auth';
import { useStreamStore } from '@/stores/stream';
import { ref } from 'vue';
import { toast } from 'vue3-toastify';

const recordings = ref([]);
const isLoading = ref(false);

const authUser = useAuthStore().authUser;
if (authUser.role !== 'interviewer') {
    router.push({ name: 'Home' });
}

const getRecordings = async () => {
    const client = useStreamStore().client;
    const authUser = useAuthStore().authUser;
    isLoading.value = true;
    try {
        const { calls } = await client.queryCalls({
            sort: [{ field: 'starts_at', direction: -1 }],
            filter_conditions: {
                starts_at: { $exists: true },
                $or: [{ created_by_user_id: authUser.uuid }, { members: { $in: [authUser.uuid] } }],
            },
        });

        const callRecordings = await Promise.all(calls.map((call) => call.queryRecordings()));
        recordings.value = callRecordings.flatMap((call) => call.recordings);
    } catch (err) {
        console.log(err);
        toast.error('Failed to load your recordings!');
    } finally {
        isLoading.value = false;
    }
};

getRecordings();
</script>
