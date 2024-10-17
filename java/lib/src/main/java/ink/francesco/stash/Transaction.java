package ink.francesco.stash;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;

public class Transaction {

    private long hnd;
    private ObjectMapper mapper = new ObjectMapper();

    protected Transaction(long hnd) {
        this.hnd = hnd;
    }

    public int exec(String sql, Map<String, Object> args) throws Exception { 
        Result res = StashLibrary.instance.stash_exec(hnd, sql, mapper.writeValueAsString(args));
        return res.obj(Integer.class);
    }

    public void commit() throws Exception {
        Result res = StashLibrary.instance.stash_commit(hnd);
        res.check();
    }
    
    public void sync() {
        Result res = StashLibrary.instance.stash_sync(hnd);
        res.check();
    }

    public void cancel() {
        Result res = StashLibrary.instance.stash_rollback(hnd);
        res.check();
    }

}
