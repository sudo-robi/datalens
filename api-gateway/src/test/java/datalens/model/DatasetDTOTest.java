package datalens.model;

import org.junit.jupiter.api.Test;

import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.*;

class DatasetDTOTest {

    @Test
    void fromEntityMapsAllFieldsCorrectly() {
        User user = User.builder()
                .id(1L).username("u").email("u@x.com").passwordHash("h").build();

        LocalDateTime now = LocalDateTime.now();
        Dataset dataset = Dataset.builder()
                .id(5L)
                .name("Sales Data")
                .description("Q4 numbers")
                .filename("sales.csv")
                .mimeType("text/csv")
                .rowCount(250L)
                .sizeBytes(4096L)
                .user(user)
                .createdAt(now)
                .updatedAt(now)
                .build();

        DatasetDTO dto = DatasetDTO.fromEntity(dataset);

        assertEquals(5L, dto.getId());
        assertEquals("Sales Data", dto.getName());
        assertEquals("Q4 numbers", dto.getDescription());
        assertEquals("sales.csv", dto.getFilename());
        assertEquals("text/csv", dto.getMimeType());
        assertEquals(250L, dto.getRowCount());
        assertEquals(4096L, dto.getSizeBytes());
        assertEquals(now, dto.getCreatedAt());
        assertEquals(now, dto.getUpdatedAt());
    }

    @Test
    void fromEntityHandlesNullDescription() {
        User user = User.builder()
                .id(1L).username("u").email("u@x.com").passwordHash("h").build();

        LocalDateTime now = LocalDateTime.now();
        Dataset dataset = Dataset.builder()
                .id(6L)
                .name("No Desc")
                .description(null)
                .filename("nd.csv")
                .mimeType("text/csv")
                .rowCount(10L)
                .sizeBytes(100L)
                .user(user)
                .createdAt(now)
                .updatedAt(now)
                .build();

        DatasetDTO dto = DatasetDTO.fromEntity(dataset);

        assertEquals(6L, dto.getId());
        assertEquals("No Desc", dto.getName());
        assertNull(dto.getDescription());
    }
}
