package datalens.model;

import org.junit.jupiter.api.Test;

import java.lang.reflect.Method;
import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.*;

class DatasetTest {

    @Test
    void builderCreatesValidDatasetWithAllFields() {
        User owner = User.builder().id(1L).username("alice").build();
        LocalDateTime now = LocalDateTime.now();

        Dataset dataset = Dataset.builder()
                .id(10L)
                .name("Sales Data")
                .description("Q4 sales")
                .filename("sales.csv")
                .storagePath("/data/sales.csv")
                .mimeType("text/csv")
                .rowCount(1000L)
                .sizeBytes(204800L)
                .user(owner)
                .createdAt(now)
                .build();

        assertEquals(10L, dataset.getId());
        assertEquals("Sales Data", dataset.getName());
        assertEquals("Q4 sales", dataset.getDescription());
        assertEquals("sales.csv", dataset.getFilename());
        assertEquals("/data/sales.csv", dataset.getStoragePath());
        assertEquals("text/csv", dataset.getMimeType());
        assertEquals(1000L, dataset.getRowCount());
        assertEquals(204800L, dataset.getSizeBytes());
        assertEquals(owner, dataset.getUser());
        assertEquals(now, dataset.getCreatedAt());
    }

    @Test
    void prePersistSetsCreatedAtAndUpdatedAt() throws Exception {
        Dataset dataset = new Dataset();
        LocalDateTime before = LocalDateTime.now();

        Method onCreate = Dataset.class.getDeclaredMethod("onCreate");
        onCreate.setAccessible(true);
        onCreate.invoke(dataset);

        LocalDateTime after = LocalDateTime.now();
        assertNotNull(dataset.getCreatedAt());
        assertNotNull(dataset.getUpdatedAt());
        assertEquals(dataset.getCreatedAt(), dataset.getUpdatedAt());
        assertFalse(dataset.getCreatedAt().isBefore(before));
        assertFalse(dataset.getCreatedAt().isAfter(after));
    }

    @Test
    void preUpdateSetsUpdatedAt() throws Exception {
        Dataset dataset = new Dataset();
        dataset.setCreatedAt(LocalDateTime.now().minusHours(1));
        LocalDateTime beforeUpdate = LocalDateTime.now();

        Method onUpdate = Dataset.class.getDeclaredMethod("onUpdate");
        onUpdate.setAccessible(true);
        onUpdate.invoke(dataset);

        LocalDateTime afterUpdate = LocalDateTime.now();
        assertNotNull(dataset.getUpdatedAt());
        assertFalse(dataset.getUpdatedAt().isBefore(beforeUpdate));
        assertFalse(dataset.getUpdatedAt().isAfter(afterUpdate));
    }

    @Test
    void noArgsConstructorCreatesEmptyDataset() {
        Dataset dataset = new Dataset();
        assertNull(dataset.getId());
        assertNull(dataset.getName());
    }
}
