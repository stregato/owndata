import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.jna.NativeLong;
import com.sun.jna.Pointer;
import com.sun.jna.Structure;

public class Result extends Structure implements Structure.ByValue {
    public Pointer ptr;
    public NativeLong len;
    public long hnd;
    public Pointer err; // Use Pointer for char*
    ObjectMapper mapper = new ObjectMapper();

    @Override
    protected List<String> getFieldOrder() {
        return Arrays.asList("ptr", "len", "hnd", "err");
    }
    public long getHnd() {
        return hnd;
    }

    public byte[] getData() {
        check();
        byte[] data =  ptr == null ? null : ptr.getByteArray(0, (int) len.longValue());
        System.out.println("getData: " + new String(data));
        return data;
    }

    public String getError() {
        return err == null ? null : err.getString(0);
    }

    public void check() {
        if (getError() != null) {
            throw new RuntimeException(getError());
        }
    }

    public<K,V> Map<K, V> map(Class<K> keyClass, Class<V> valueClass) throws IOException {
        check();
        return mapper.readValue(getData(), mapper.getTypeFactory().constructMapType(Map.class, keyClass, valueClass));
    }

    public<T> List<T> list(Class<T> clazz)  throws IOException {
        check();
        byte[] data = getData();
        List<T> list = mapper.readValue(data, mapper.getTypeFactory()
        .constructCollectionType(List.class, clazz));
        return list == null ? new ArrayList<>() : list;
    }

    public<T> T obj(Class<T> clazz) throws IOException {
        check();
        return mapper.readValue(getData(), clazz);
    }

    @Override
    public String toString() {
        byte[] data = getData();
        String error = getError();

        String res = "Result{";
        if (hnd != 0) {
            res += "hnd=" + hnd+ ", ";
        }
        if (data != null) {
            res += "data=" + new String(data) + ", ";
        }
        if (error != null) {
            res += "error='" + error + '\'';
        }
        return res + '}';
    }
}


// public class Result {
//     private final long hnd;
//     private final byte[] data;
//     private final String error;

//     public Result(long hnd, byte[] data, String error) {
//         this.hnd = hnd;
//         this.data = data;
//         this.error = error;
//     }

//     public long getHnd() {
//         return hnd;
//     }

//     public byte[] getData() {
//         return data;
//     }

//     public String getError() {
//         return error;
//     }

//     public void check() {
//         if (error != null) {
//             throw new RuntimeException(error);
//         }
//     }

//     public Map<String, Object> map() throws IOException {
//         check();
//         if (data == null) {
//             return null;
//         }
//         return new ObjectMapper().readValue(data, Map.class);
//     }

//     public List<Object> list()  throws IOException {
//         check();
//         if (data == null) {
//             return null;
//         }
//             return new ObjectMapper().readValue(data, List.class);
//     }

//     @Override
//     public String toString() {
//         return "Result{" +
//                 "hnd=" + hnd +
//                 ", data=" + (data != null ? new String(data) : "null") +
//                 ", error='" + error + '\'' +
//                 '}';
//     }
// }

