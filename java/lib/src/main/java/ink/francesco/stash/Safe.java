package ink.francesco.stash;

import java.util.Set;
import java.util.Map;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;




public class Safe {

    public static String usrGroup = "usr";
    public static String adminGroup = "admin";

    static public class Config {
    @JsonProperty
    long quota;
    @JsonProperty
    String description;

    String toJson() throws JsonProcessingException {
        ObjectMapper mapper = new ObjectMapper();
        return mapper.writeValueAsString(this);
    }
}

    static public enum GroupChange {
        Add(1), Remove(2);
        
        private final int value;
        GroupChange(int value) {
            this.value = value;
        }
    }

    long hnd;
    ObjectMapper mapper = new ObjectMapper();

    static public Safe create(DB db, Identity identity, String url, 
                Config config ) throws Exception {
        Result r = StashLibrary.instance.stash_createSafe(db.hnd, identity.toJson(), url, config.toJson());
        r.check();
        r.check();
        return new Safe(r.hnd);
    }

    static public Safe open(DB db, Identity identity, String url) throws Exception {
        Result r = StashLibrary.instance.stash_openSafe(db.hnd, identity.toJson(), url);
        r.check();
        return new Safe(r.hnd);
    }

    public void close()  {
        StashLibrary.instance.stash_closeSafe(hnd);
    }

    private Safe(long hnd) {
        this.hnd = hnd;
    }

    public Set<String> updateGroup(String groupName, GroupChange change, String[] users) throws Exception {
        Result r = StashLibrary.instance.stash_updateGroup(hnd, groupName, change.value, mapper.writeValueAsString(users));
        r.check();
        return r.map(String.class, Boolean.class).keySet();
    }

    public Set<String> getGroups() throws Exception {
        Result r = StashLibrary.instance.stash_getGroups(hnd);
        r.check();
        return r.map(String.class, Boolean.class).keySet();
    }

    public Set<String> getKeys(String groupName, long expectedMinimumLenght) throws Exception {
        Result r = StashLibrary.instance.stash_getKeys(hnd, groupName, expectedMinimumLenght);
        r.check();
        return r.map(String.class, Boolean.class).keySet();
    }

    public FS openFS() throws Exception {
        Result r = StashLibrary.instance.stash_openFS(hnd);
        r.check();
        return new FS(r.hnd);
    }

    public Database openDatabase(String groupName, Map<String,String> ddls) throws Exception {
        Result r = StashLibrary.instance.stash_openDatabase(hnd, groupName, mapper.writeValueAsString(ddls));
        r.check();
        return new Database(r.hnd);
    }

    public Comm openComm() throws Exception {
        Result r = StashLibrary.instance.stash_openComm(hnd);
        r.check();
        return new Comm(r.hnd);
    }
}
