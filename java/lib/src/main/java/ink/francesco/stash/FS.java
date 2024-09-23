package ink.francesco.stash;

import java.util.Date;
import java.util.Map;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;

public class FS {
    long hnd;
    ObjectMapper mapper = new ObjectMapper();


    static public class ListOptions {
        @JsonProperty
        Date after;
        @JsonProperty
        Date before;
        @JsonProperty
        String orderBy;
        @JsonProperty
        boolean reverse;
        @JsonProperty
        int limit;
        @JsonProperty
        int offset;
        @JsonProperty
        String prefix;
        @JsonProperty
        String suffix;
        @JsonProperty
        String tag;
    }

    static public class File {
        @JsonProperty
        public long id;
        @JsonProperty
        public String dir;
        @JsonProperty
        public String name;
        @JsonProperty
        public boolean isDir;
        @JsonProperty
        public String groupName;
        @JsonProperty
        public String creator;
        @JsonProperty
        public int size;
        @JsonProperty
        public Date modTime;
        @JsonProperty
        public List<String> tags;
        @JsonProperty
        public Map<String, Object> attributes;
        @JsonProperty
        public String localCopy;
        @JsonProperty
        public Date copyTime;
        @JsonProperty
        public byte[] encryptionKey;
    }


    static public class PutOptions {
        @JsonProperty
        long id;
        @JsonProperty
        boolean async;
        @JsonProperty
        boolean deleteSrc;
        @JsonProperty
        String groupName;
        @JsonProperty
        List<String> tags;
        @JsonProperty
        String attributes;
    }

    static public class GetOptions {
        @JsonProperty
        boolean async;
    }

    public FS(long hnd) {
        this.hnd = hnd;
    }

    public void close() {
        StashLibrary.instance.stash_closeFS(hnd);
    }

    public List<File> list(String path, ListOptions options) throws Exception {
        Result r = StashLibrary.instance.stash_list(hnd, path, mapper.writeValueAsString(options));
        return r.list(File.class);
    }

    public File stat(String path) throws Exception {
        Result r = StashLibrary.instance.stash_stat(hnd, path);
        return r.obj(File.class);
    }

    public File putFile(String dest, String src, PutOptions options) throws Exception {
        Result r = StashLibrary.instance.stash_putFile(hnd, dest, src, mapper.writeValueAsString(options));
        return r.obj(File.class);
    }

    public File putData(String dest, byte[] data, PutOptions options) throws Exception {
        Result r = StashLibrary.instance.stash_putData(hnd, dest, new Data(data), mapper.writeValueAsString(options));
        return r.obj(File.class);
    }

    public File getFile(String src, String dest, GetOptions options) throws Exception {
        Result r = StashLibrary.instance.stash_getFile(hnd, src, dest, mapper.writeValueAsString(options));
        return r.obj(File.class);
    }

    public byte[] getData(String src, GetOptions options) throws Exception {
        Result r = StashLibrary.instance.stash_getData(hnd, src, mapper.writeValueAsString(options));
        return r.getData();
    }

    public void delete(String path) throws Exception {
        StashLibrary.instance.stash_delete(hnd, path).check();
    }

    public File rename(String oldPath, String newPath) throws Exception {
        Result r = StashLibrary.instance.stash_rename(hnd, oldPath, newPath);
        return r.obj(File.class);
    }

    
}
