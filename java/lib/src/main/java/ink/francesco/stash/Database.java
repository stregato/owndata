package ink.francesco.stash;
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

    public Transaction transaction() {
        Result res = StashLibrary.instance.stash_transaction(hnd);
        res.check();
        return new Transaction(res.hnd);
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
        Result res = StashLibrary.instance.stash_rollback(hnd);
        res.check();
    }

}
