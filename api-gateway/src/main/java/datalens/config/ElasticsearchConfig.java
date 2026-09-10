package datalens.config;

import org.apache.http.HttpHost;
import org.elasticsearch.client.RestClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class ElasticsearchConfig {

    @Value("${elasticsearch.url}")
    private String elasticsearchUrl;

    @Bean
    public RestClient elasticsearchClient() {
        return RestClient.builder(HttpHost.create(elasticsearchUrl)).build();
    }
}
