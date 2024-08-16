import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;

public class Database {

    public class Rows {
        private long hnd;

        private Rows(long hnd) {
            this.hnd = hnd;
        }

        public Map<String, Object> next() throws Exception {
            Result result = StashLibrary.instance.stash_nextRow(hnd);
            return result.map(String.class, Object.class);
        }

        public void close() {
            StashLibrary.instance.stash_closeRows(hnd);
        }
    }

    private long hnd;
    private ObjectMapper mapper = new ObjectMapper();

    protected Database(long hnd) {
        this.hnd = hnd;
    }

    public void close() {
        StashLibrary.instance.stash_closeDatabase(hnd);
    }

    public int exec(String sql, Map<String, Object> args) throws Exception { 
        Result res = StashLibrary.instance.stash_exec(hnd, sql, mapper.writeValueAsString(args));
        return res.obj(Integer.class);
    }

    public Rows query(String sql, Map<String, Object> args) throws Exception {
        Result res = StashLibrary.instance.stash_query(hnd, sql, mapper.writeValueAsString(args));
        res.check();
        return new Rows(res.hnd);
    }
    
    public void sync() {
        Result res = StashLibrary.instance.stash_sync(hnd);
        res.check();
    }

    public void cancel() {
        Result res = StashLibrary.instance.stash_cancel(hnd);
        res.check();
    }

}
