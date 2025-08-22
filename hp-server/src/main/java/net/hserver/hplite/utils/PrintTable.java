package net.hserver.hplite.utils;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * @author vdi.zhoul
 * @description: 控制台打印表格工具类 支持全角和半角两种模式
 * @date 2023/1/31 16:38
 */
public class PrintTable {
    private static final char HALF_ROW_LINE = '-';
    private static final char FULL_ROW_LINE = '－';
    private static final char COLUMN_LINE = '|';
    private static final char CORNER = '+';
    //半角模式空格的unicode码
    private static final char HALF_SPACE = '\u0020';
    //全角模式空格的unicode码
    private static final char FULL_SPACE = '\u3000';
    private static final char LF = '\n';
    private static final char SPACE = ' ';
    private static final String EMPTY = "";

    //全角模式
    private boolean sbcMode = true;

    public static PrintTable create() {
        return new PrintTable();
    }

    //表格头信息
    private final List<List<String>> headerList = new ArrayList<>();

    //表格体信息
    private final List<List<String>> bodyList = new ArrayList<>();

    //每列最大字符个数
    private List<Integer> columnCharCount;

    /**
     * 设置是否使用全角模式
     *
     * @param sbcMode 是否全角模式
     */
    public PrintTable setSbcMode(boolean sbcMode) {
        this.sbcMode = sbcMode;
        return this;
    }

    /**
     * 添加表头
     *
     * @param titles 列名
     */
    public PrintTable addHeader(String... titles) {
        if (columnCharCount == null) {
            columnCharCount = new ArrayList<>(Collections.nCopies(titles.length, 0));
        }
        List<String> headers = new ArrayList<>();
        fillColumns(headers, titles);
        headerList.add(headers);
        return this;
    }

    /**
     * 添加表体
     *
     * @param values 列值
     */
    public PrintTable addBody(String... values) {
        List<String> bodies = new ArrayList<>();
        bodyList.add(bodies);
        fillColumns(bodies, values);
        return this;
    }

    /**
     * 填充表头或者表体
     *
     * @param columns 被填充列表
     * @param values  填充值
     */
    private void fillColumns(List<String> columns, String[] values) {
        for (int i = 0; i < values.length; i++) {
            String column = values[i];
            if (sbcMode) {
                column = toSbc(column);
            }
            columns.add(column);
            int width = column.length();
            if (!sbcMode) {
                int sbcCount = nonSbcCount(column);
                width = (width - sbcCount) * 2 + sbcCount;
            }
            if (width > columnCharCount.get(i)) {
                columnCharCount.set(i, width);
            }
        }
    }

    /**
     * 获取表格字符串
     */
    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append("\n");
        fillBorder(sb);
        fillRows(sb, headerList);
        fillBorder(sb);
        fillRows(sb, bodyList);
        fillBorder(sb);
        return sb.toString();
    }

    /**
     * 填充表头或者表体信息（多行）
     *
     * @param sb   内容
     * @param list 表头列表或者表体列表
     */
    private void fillRows(StringBuilder sb, List<List<String>> list) {
        for (List<String> row : list) {
            sb.append(COLUMN_LINE);
            fillRow(sb, row);
            sb.append(LF);
        }
    }

    /**
     * 填充一行数据
     *
     * @param sb  内容
     * @param row 一行数据
     */
    private void fillRow(StringBuilder sb, List<String> row) {
        final int size = row.size();
        String value;
        for (int i = 0; i < size; i++) {
            value = row.get(i);
            sb.append(sbcMode ? FULL_SPACE : HALF_SPACE);
            sb.append(value);
            final int length = value.length();
            final int sbcCount = nonSbcCount(value);
            if (sbcMode && sbcCount % 2 == 1) {
                sb.append(SPACE);
            }
            sb.append(sbcMode ? FULL_SPACE : HALF_SPACE);
            int maxLength = columnCharCount.get(i);
            int doubleNum = 2;
            if (sbcMode) {
                for (int j = 0; j < (maxLength - length + (sbcCount / doubleNum)); j++) {
                    sb.append(FULL_SPACE);
                }
            } else {
                for (int j = 0; j < (maxLength - ((length - sbcCount) * doubleNum + sbcCount)); j++) {
                    sb.append(HALF_SPACE);
                }
            }
            sb.append(COLUMN_LINE);
        }
    }

    /**
     * 填充边框
     *
     * @param sb StringBuilder
     */
    private void fillBorder(StringBuilder sb) {
        sb.append(CORNER);
        for (Integer width : columnCharCount) {
            sb.append(sbcMode ? repeat(FULL_ROW_LINE, width + 2) : repeat(HALF_ROW_LINE, width + 2));
            sb.append(CORNER);
        }
        sb.append(LF);
    }

    /**
     * 打印到控制台
     */
    public void print() {
        System.out.print(this);
    }

    /**
     * 半角字符数量<br/>
     * 英文字母、数字键、符号键
     *
     * @param value 字符串
     */
    private int nonSbcCount(String value) {
        int count = 0;
        for (int i = 0; i < value.length(); i++) {
            if (value.charAt(i) < '\177') {
                count++;
            }
        }
        return count;
    }

    /**
     * 重复字符
     *
     * @param c     字符
     * @param count 重复次数
     */
    public static String repeat(char c, int count) {
        if (count <= 0) {
            return EMPTY;
        }

        char[] result = new char[count];
        for (int i = 0; i < count; i++) {
            result[i] = c;
        }
        return new String(result);
    }

    /**
     * 转成全角字符
     *
     * @param input 字符
     */
    public static String toSbc(String input) {
        final char[] c = input.toCharArray();
        for (int i = 0; i < c.length; i++) {
            if (c[i] == ' ') {
                c[i] = '\u3000';
            } else if (c[i] < '\177') {
                c[i] = (char) (c[i] + 65248);

            }
        }
        return new String(c);
    }

    public static void main(String[] args) {
        PrintTable printTable1 = PrintTable.create();
        PrintTable printTable2 = PrintTable.create();

        printTable1.setSbcMode(true);
        printTable(printTable1);

        System.out.println("====全角模式效果=====");
        printTable1.print();
        printTable2.setSbcMode(false);
        printTable(printTable2);

        System.out.println("====半角模式效果=====");
        printTable2.print();
    }

    private static void printTable(PrintTable printTable1) {
        printTable1.addHeader("姓名", "备注", "住址", "email");
        printTable1.addBody("小明", "20", "北京", "123@qq.com");
        printTable1.addBody("小红", "15", "上海", "123456@qq.com");
        printTable1.addBody("小李", "40", "深圳", "123789@qq.com");
        printTable1.addBody("小王", "36", "广州", "1111111@qq.com");
    }
}

