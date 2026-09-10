package datalens.model;

import lombok.Data;
import lombok.Builder;
import java.time.LocalDateTime;

@Data @Builder
public class DatasetDTO {
    private Long id;
    private String name;
    private String description;
    private String filename;
    private String mimeType;
    private Long rowCount;
    private Long sizeBytes;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    public static DatasetDTO fromEntity(datalens.model.Dataset dataset) {
        return DatasetDTO.builder()
                .id(dataset.getId())
                .name(dataset.getName())
                .description(dataset.getDescription())
                .filename(dataset.getFilename())
                .mimeType(dataset.getMimeType())
                .rowCount(dataset.getRowCount())
                .sizeBytes(dataset.getSizeBytes())
                .createdAt(dataset.getCreatedAt())
                .updatedAt(dataset.getUpdatedAt())
                .build();
    }
}
